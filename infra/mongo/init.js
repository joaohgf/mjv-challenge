try {
  rs.status();
} catch (error) {
  rs.initiate({ _id: 'rs0', members: [{ _id: 0, host: 'mongo:27017' }] });
}

while (!db.hello().isWritablePrimary) {
  sleep(100);
}

const orders = db.getSiblingDB('orders');

orders.orders.createIndex(
  { order_id: 1 },
  {
    unique: true,
    partialFilterExpression: { order_id: { $type: 'string' } },
  },
);

orders.runCommand({
  createIndexes: 'outbox',
  indexes: [
    {
      key: { event_id: 1 },
      name: 'event_id_1',
      unique: true,
      partialFilterExpression: { event_id: { $type: 'string' } },
    },
    { key: { status: 1, created_at: 1 }, name: 'status_1_created_at_1' },
    { key: { status: 1, locked_until: 1 }, name: 'status_1_locked_until_1' },
  ],
});
