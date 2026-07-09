#!/usr/bin/env node

import process from "node:process";

const uri = process.env.MONGODB_URI || process.env.MONGOOSE_CONNECTION_STRING;
const dbName = process.env.MONGODB_DATABASE;

if (!uri || !dbName) {
  console.error("missing required env: MONGODB_URI or MONGOOSE_CONNECTION_STRING, and MONGODB_DATABASE");
  process.exit(2);
}

let mongodb;
try {
  mongodb = await import("mongodb");
} catch (err) {
  console.error("missing dependency: install mongodb driver before running this audit tool");
  console.error("example: npm install --no-save mongodb");
  process.exit(2);
}

const { MongoClient, ReadPreference, BSON } = mongodb;

const expectedCollections = new Map([
  ["numbers", { duplicateKeys: [["guild"]] }],
  ["all_use_counts", { duplicateKeys: [["slashcommand_name"]] }],
  ["ann_all_sets", { duplicateKeys: [["guild", "announcement_id"]] }],
  ["birthdays", { duplicateKeys: [["guild", "user"]] }],
  ["birthday_sets", { duplicateKeys: [["guild"]] }],
  ["btns", { duplicateKeys: [["guild", "number"]] }],
  ["chats", { duplicateKeys: [["guild"]] }],
  ["chat_roles", { duplicateKeys: [["guild", "leavel", "role"]] }],
  ["chatgpts", { duplicateKeys: [["guild"]] }],
  ["chatgpt_gets", { duplicateKeys: [["guild"]] }],
  ["codes", { duplicateKeys: [["code"]] }],
  ["coins", { duplicateKeys: [["guild", "member"]] }],
  ["create_hours", { duplicateKeys: [["guild"]] }],
  ["cron_sets", { duplicateKeys: [["guild", "id"]] }],
  ["errors_sets", { duplicateKeys: [["guild"]] }],
  ["ghps", { duplicateKeys: [["guild", "commodity_id"]] }],
  ["gifts", { duplicateKeys: [["guild", "gift_name"]] }],
  ["gift_changes", { duplicateKeys: [["guild"]] }],
  ["good_webs", { duplicateKeys: [["guild"]] }],
  ["guilds", { duplicateKeys: [["guild"]] }],
  ["join_messages", { duplicateKeys: [["guild"], ["enable"]] }],
  ["join_roles", { duplicateKeys: [["guild", "role"]] }],
  ["leave_messages", { duplicateKeys: [["guild"]] }],
  ["lock_channels", { duplicateKeys: [["guild", "channel_id"]] }],
  ["loggings", { duplicateKeys: [["guild"]] }],
  ["lotters", { duplicateKeys: [["guild", "id"]] }],
  ["message_reactions", { duplicateKeys: [["guild", "message", "react"]] }],
  ["not_a_good_webs", { duplicateKeys: [["web"]] }],
  ["polls", { duplicateKeys: [["guild", "messageid"]] }],
  ["role_numbers", { duplicateKeys: [["guild", "role"], ["guild", "channel"]] }],
  ["sign_lists", { duplicateKeys: [["guild", "member"]] }],
  ["suports", { duplicateKeys: [["support_id"]] }],
  ["systems", { duplicateKeys: [] }],
  ["text_xps", { duplicateKeys: [["guild", "member"]] }],
  ["text_xp_channels", { duplicateKeys: [["guild"]] }],
  ["tickets", { duplicateKeys: [["guild"]] }],
  ["verifications", { duplicateKeys: [["guild"]] }],
  ["voice_channels", { duplicateKeys: [["guild", "ticket_channel"]] }],
  ["voice_channel_ids", { duplicateKeys: [["guild", "channel_id"]] }],
  ["voice_roles", { duplicateKeys: [["guild", "leavel", "role"]] }],
  ["voice_xps", { duplicateKeys: [["guild", "member"]] }],
  ["voice_xp_channels", { duplicateKeys: [["guild"]] }],
  ["votes", { duplicateKeys: [["guild", "Number"], ["guild", "member"]] }],
  ["warndbs", { duplicateKeys: [["guild", "user"]] }],
  ["work_sets", { duplicateKeys: [["guild"]] }],
  ["work_somethings", { duplicateKeys: [["guild", "name"]] }],
  ["work_users", { duplicateKeys: [["guild", "user"]] }],
]);

const requiredFields = {
  coins: ["guild", "member"],
  text_xps: ["guild", "member"],
  voice_xps: ["guild", "member"],
  work_users: ["guild", "user"],
  polls: ["guild", "messageid"],
  cron_sets: ["guild", "id", "cron", "channel"],
  join_messages: ["guild"],
  loggings: ["guild"],
};

const impossibleChecks = {
  coins: [
    { field: "coin", min: 0 },
    { field: "today", min: 0 },
  ],
  text_xps: [
    { field: "xp", min: 0 },
    { field: "leavel", min: 0 },
  ],
  voice_xps: [
    { field: "xp", min: 0 },
    { field: "leavel", min: 0 },
  ],
  gifts: [
    { field: "gift_count", min: 0 },
    { field: "gift_chence", min: 0 },
  ],
  work_users: [
    { field: "energi", min: 0 },
    { field: "get_coin", min: 0 },
  ],
};

function bsonType(value) {
  if (value === null) return "null";
  if (Array.isArray(value)) return "array";
  if (value instanceof Date) return "date";
  return typeof value;
}

function flattenShape(doc, prefix = "", out = {}) {
  for (const [key, value] of Object.entries(doc)) {
    const name = prefix ? `${prefix}.${key}` : key;
    out[name] = bsonType(value);
    if (value && typeof value === "object" && !Array.isArray(value) && !(value instanceof Date) && key !== "_id") {
      flattenShape(value, name, out);
    }
  }
  return out;
}

function redactURI(input) {
  return input.replace(/\/\/([^:@/]+):([^@/]+)@/, "//<redacted-user>:<redacted-pass>@");
}

async function duplicateAudit(collection, keys) {
  if (keys.length === 0) return [];
  const id = {};
  const missingExpr = [];
  for (const key of keys) {
    id[key] = `$${key}`;
    missingExpr.push({ $eq: [{ $type: `$${key}` }, "missing"] });
  }
  return collection
    .aggregate([
      { $group: { _id: id, count: { $sum: 1 }, sampleIds: { $push: "$_id" } } },
      { $match: { count: { $gt: 1 }, $expr: { $not: [{ $or: missingExpr }] } } },
      { $project: { _id: 1, count: 1, sampleIds: { $slice: ["$sampleIds", 5] } } },
      { $limit: 20 },
    ])
    .toArray();
}

async function typeAudit(collection, field) {
  return collection
    .aggregate([
      { $project: { type: { $type: `$${field}` } } },
      { $group: { _id: "$type", count: { $sum: 1 } } },
      { $sort: { count: -1 } },
    ])
    .toArray();
}

async function missingFieldCount(collection, field) {
  return collection.countDocuments({ [field]: { $exists: false } });
}

async function impossibleCount(collection, check) {
  return collection.countDocuments({
    [check.field]: { $exists: true, $type: "number", $lt: check.min },
  });
}

const client = new MongoClient(uri, {
  readPreference: ReadPreference.SECONDARY_PREFERRED,
  retryWrites: false,
  appName: "mhcat-readonly-audit",
});

try {
  await client.connect();
  const db = client.db(dbName);
  const liveCollections = await db.listCollections({}, { nameOnly: true }).toArray();
  const liveNames = liveCollections.map((c) => c.name).sort();
  const expectedNames = [...expectedCollections.keys()].sort();
  const unexpected = liveNames.filter((name) => !expectedCollections.has(name));
  const missing = expectedNames.filter((name) => !liveNames.includes(name));

  const report = {
    generatedAt: new Date().toISOString(),
    database: dbName,
    uri: redactURI(uri),
    readOnly: true,
    expectedCollectionCount: expectedNames.length,
    liveCollectionCount: liveNames.length,
    missingExpectedCollections: missing,
    collectionsNotRepresentedByMongoose: unexpected,
    collections: {},
  };

  for (const name of liveNames) {
    const collection = db.collection(name);
    const expected = expectedCollections.get(name);
    const indexes = await collection.indexes();
    const count = await collection.estimatedDocumentCount();
    const samples = await collection.find({}, { limit: 5 }).toArray();
    const sampleShapes = samples.map((doc) => flattenShape(doc));
    const largeDocs = await collection
      .aggregate([
        { $project: { size: { $bsonSize: "$$ROOT" } } },
        { $match: { size: { $gt: 1024 * 256 } } },
        { $sort: { size: -1 } },
        { $limit: 10 },
      ])
      .toArray();

    const missingRequired = {};
    for (const field of requiredFields[name] || []) {
      missingRequired[field] = await missingFieldCount(collection, field);
    }

    const mixedTypes = {};
    const fields = new Set();
    for (const shape of sampleShapes) {
      for (const field of Object.keys(shape)) fields.add(field);
    }
    for (const field of [...fields].filter((f) => f !== "_id").slice(0, 50)) {
      mixedTypes[field] = await typeAudit(collection, field);
    }

    const duplicates = {};
    for (const keys of expected?.duplicateKeys || []) {
      duplicates[keys.join("+")] = await duplicateAudit(collection, keys);
    }

    const impossibleValues = {};
    for (const check of impossibleChecks[name] || []) {
      impossibleValues[check.field] = await impossibleCount(collection, check);
    }

    report.collections[name] = {
      representedByMongoose: Boolean(expected),
      count,
      indexes,
      sampleShapes,
      missingRequired,
      mixedTypes,
      duplicateLogicalKeys: duplicates,
      impossibleValues,
      largeDocumentsOver256KiB: largeDocs,
    };
  }

  console.log(JSON.stringify(report, null, 2));
} finally {
  await client.close();
}
