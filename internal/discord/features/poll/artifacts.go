package poll

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/responses"
)

const (
	pollChartFileName = "file.jpg"
	pollTextFileName  = "discord.txt"
	pollXLSXFileName  = "poll_info.xlsx"
)

var pollChartColors = []color.RGBA{
	{R: 255, G: 99, B: 132, A: 255},
	{R: 54, G: 162, B: 235, A: 255},
	{R: 255, G: 205, B: 86, A: 255},
	{R: 255, G: 88, B: 8, A: 255},
	{R: 40, G: 255, B: 41, A: 255},
	{R: 2, G: 255, B: 255, A: 255},
	{R: 159, G: 52, B: 255, A: 255},
	{R: 255, G: 211, B: 8, A: 255},
	{R: 2, G: 247, B: 142, A: 255},
	{R: 174, G: 87, B: 164, A: 255},
	{R: 90, G: 90, B: 173, A: 255},
	{R: 79, G: 157, B: 157, A: 255},
	{R: 155, G: 255, B: 2, A: 255},
	{R: 148, G: 148, B: 73, A: 255},
	{R: 249, G: 249, B: 0, A: 255},
	{R: 152, G: 75, B: 75, A: 255},
	{R: 142, G: 142, B: 142, A: 255},
	{R: 255, G: 160, B: 67, A: 255},
	{R: 255, G: 68, B: 255, A: 255},
}

func pollResultMessage(ctx context.Context, poll domain.Poll, members ports.DiscordGuildMemberReader) responses.Message {
	message := pollResultEmbedMessage(poll)
	message.Embeds[0].Image = &responses.EmbedImage{URL: "attachment://" + pollChartFileName}
	message.Files = []responses.File{
		{Name: pollChartFileName, ContentType: "image/jpeg", Data: buildPollChartJPEG(poll)},
		{Name: pollTextFileName, ContentType: "text/plain; charset=utf-8", Data: []byte(buildPollTextExport(ctx, poll, members))},
	}
	return message
}

func pollExcelMessage(ctx context.Context, poll domain.Poll, members ports.DiscordGuildMemberReader) (responses.Message, error) {
	data, err := buildPollXLSX(ctx, poll, members)
	if err != nil {
		return responses.Message{}, err
	}
	return responses.Message{
		Content: "<:sheets:1023972957330100324> | **以下是該投票的excel表格!**",
		Files: []responses.File{{
			Name:        pollXLSXFileName,
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			Data:        data,
		}},
		Ephemeral: true,
	}, nil
}

func buildPollTextExport(ctx context.Context, poll domain.Poll, members ports.DiscordGuildMemberReader) string {
	tags := pollMemberTags(ctx, poll, members)
	lines := make([]string, 0, len(poll.Votes))
	for _, vote := range poll.Votes {
		userID := vote.UserID
		userTag := pollMemberTag(vote.UserID, tags)
		if poll.Anonymous {
			userID = "該投票為匿名"
			userTag = "該投票為匿名"
		}
		lines = append(lines, fmt.Sprintf("使用者id:%s|使用者名稱:%s|使用者投給的選項:%s|投票時間:%s", userID, userTag, vote.Choice, legacyPollTime(vote.Time)))
	}
	return strings.Join(lines, "\n")
}

func buildPollXLSX(ctx context.Context, poll domain.Poll, members ports.DiscordGuildMemberReader) ([]byte, error) {
	tags := pollMemberTags(ctx, poll, members)
	rows := [][]string{{"使用者ID", "使用者dc內名稱", "使用者投給的選項", "投票時間"}}
	for _, vote := range poll.Votes {
		rows = append(rows, []string{
			vote.UserID,
			pollMemberTag(vote.UserID, tags),
			vote.Choice,
			legacyPollTime(vote.Time),
		})
	}
	return buildXLSX(rows)
}

func pollMemberTag(userID string, tags map[string]string) string {
	if tags == nil {
		return "使用者已退出伺服器!"
	}
	tag := tags[userID]
	if strings.TrimSpace(tag) == "" {
		return "使用者已退出伺服器!"
	}
	return tag
}

func pollMemberTags(ctx context.Context, poll domain.Poll, members ports.DiscordGuildMemberReader) map[string]string {
	if poll.Anonymous || members == nil {
		return nil
	}
	seen := make(map[string]struct{}, len(poll.Votes))
	ids := make([]string, 0, len(poll.Votes))
	for _, vote := range poll.Votes {
		if vote.UserID == "" {
			continue
		}
		if _, ok := seen[vote.UserID]; ok {
			continue
		}
		seen[vote.UserID] = struct{}{}
		ids = append(ids, vote.UserID)
	}
	tags, err := members.MemberTags(ctx, poll.GuildID, ids)
	if err != nil {
		return nil
	}
	return tags
}

func legacyPollTime(raw string) string {
	millis, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return raw
	}
	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		loc = time.FixedZone("Asia/Taipei", 8*60*60)
	}
	return time.UnixMilli(millis).In(loc).Format("2006/01/02 15:04:05 台北標準時間")
}

func buildPollChartJPEG(poll domain.Poll) []byte {
	const size = 500
	img := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 47, G: 49, B: 54, A: 255}}, image.Point{}, draw.Src)
	total := len(poll.Votes)
	if total == 0 {
		return encodeJPEG(img)
	}
	counts := make([]int, len(poll.Choices))
	for index, choice := range poll.Choices {
		counts[index] = poll.CountChoice(choice)
	}
	start := -math.Pi / 2
	centerX, centerY := 250.0, 245.0
	radius := 160.0
	for index, count := range counts {
		if count == 0 {
			continue
		}
		angle := float64(count) / float64(total) * math.Pi * 2
		fillPie(img, centerX, centerY, radius, start, start+angle, pollChartColors[index%len(pollChartColors)])
		start += angle
	}
	drawLegend(img, counts)
	return encodeJPEG(img)
}

func fillPie(img *image.RGBA, centerX float64, centerY float64, radius float64, start float64, end float64, fill color.RGBA) {
	minX := int(centerX - radius)
	maxX := int(centerX + radius)
	minY := int(centerY - radius)
	maxY := int(centerY + radius)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			if dx*dx+dy*dy > radius*radius {
				continue
			}
			angle := math.Atan2(dy, dx)
			if angle < -math.Pi/2 {
				angle += math.Pi * 2
			}
			if angle >= start && angle < end {
				img.SetRGBA(x, y, fill)
			}
		}
	}
}

func drawLegend(img *image.RGBA, counts []int) {
	for index, count := range counts {
		if count == 0 {
			continue
		}
		x := 28 + (index%10)*46
		y := 430 + (index/10)*24
		draw.Draw(img, image.Rect(x, y, x+22, y+16), &image.Uniform{C: pollChartColors[index%len(pollChartColors)]}, image.Point{}, draw.Src)
	}
}

func encodeJPEG(img image.Image) []byte {
	var out bytes.Buffer
	if err := jpeg.Encode(&out, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil
	}
	return out.Bytes()
}

func buildXLSX(rows [][]string) ([]byte, error) {
	var out bytes.Buffer
	archive := zip.NewWriter(&out)
	files := map[string]string{
		"[Content_Types].xml":        contentTypesXML,
		"_rels/.rels":                rootRelsXML,
		"xl/workbook.xml":            workbookXML,
		"xl/_rels/workbook.xml.rels": workbookRelsXML,
		"xl/styles.xml":              stylesXML,
		"xl/worksheets/sheet1.xml":   worksheetXML(rows),
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml", "xl/worksheets/sheet1.xml"} {
		writer, err := archive.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write([]byte(files[name])); err != nil {
			return nil, err
		}
	}
	if err := archive.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func worksheetXML(rows [][]string) string {
	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`)
	builder.WriteString(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetData>`)
	for rowIndex, row := range rows {
		builder.WriteString(fmt.Sprintf(`<row r="%d">`, rowIndex+1))
		for colIndex, value := range row {
			style := ""
			if rowIndex == 0 {
				style = ` s="1"`
			}
			builder.WriteString(fmt.Sprintf(`<c r="%s%d" t="inlineStr"%s><is><t>%s</t></is></c>`, excelColumn(colIndex+1), rowIndex+1, style, xmlEscape(value)))
		}
		builder.WriteString(`</row>`)
	}
	builder.WriteString(`</sheetData></worksheet>`)
	return builder.String()
}

func excelColumn(index int) string {
	var out []byte
	for index > 0 {
		index--
		out = append([]byte{byte('A' + index%26)}, out...)
		index /= 26
	}
	return string(out)
}

func xmlEscape(value string) string {
	var out bytes.Buffer
	_ = xml.EscapeText(&out, []byte(value))
	return out.String()
}

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Default Extension="xml" ContentType="application/xml"/>
  <Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/>
  <Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/>
  <Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/>
</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/>
</Relationships>`

const workbookXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships">
  <sheets><sheet name="poll_info" sheetId="1" r:id="rId1"/></sheets>
</workbook>`

const workbookRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/>
  <Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>
</Relationships>`

const stylesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main">
  <fonts count="2"><font/><font><b/></font></fonts>
  <fills count="1"><fill><patternFill patternType="none"/></fill></fills>
  <borders count="1"><border/></borders>
  <cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs>
  <cellXfs count="2"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="0" borderId="0" xfId="0" applyFont="1"/></cellXfs>
</styleSheet>`
