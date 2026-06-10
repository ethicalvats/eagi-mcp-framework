import { defineResource } from '@eagi/sdk';

export default defineResource({
  uri: 'tasks://{id}',
  name: 'Task Details',
  description: 'Detailed info for a specific task.',
  mimeType: 'application/json',
  handler: async (params, context) => {
    const { id } = params;
    return JSON.stringify({
      id,
      title: `Task ${id}`,
      status: 'pending',
      updatedAt: new Date().toISOString()
    });
  }
});
