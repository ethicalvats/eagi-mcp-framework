import { Server } from '@modelcontextprotocol/sdk/server/index.js';
import {
  ListToolsRequestSchema,
  CallToolRequestSchema,
  ListResourcesRequestSchema,
  ReadResourceRequestSchema,
  ListPromptsRequestSchema,
  GetPromptRequestSchema,
} from '@modelcontextprotocol/sdk/types.js';
import { zodToJsonSchema } from 'zod-to-json-schema';
import type { EagiConfig, DomainRegistry, HookEngine } from '../index';
import { authMiddleware } from '../middleware/auth';
import { auditMiddleware } from '../middleware/audit';
import { approvalMiddleware } from '../middleware/approval';

export class EagiServerBuilder {
  constructor(
    private config: EagiConfig,
    private registry: DomainRegistry,
    private hooks: HookEngine,
    private servicesMap: Record<string, any>
  ) {}

  build(): Server {
    const server = new Server(
      { name: this.config.name, version: this.config.version },
      { capabilities: { tools: {}, resources: {}, prompts: {} } }
    );

    this.registerTools(server);
    this.registerResources(server);
    this.registerPrompts(server);

    return server;
  }

  private registerTools(server: Server) {
    server.setRequestHandler(ListToolsRequestSchema, async () => {
      // Identity projection happens at the gateway level, or here via middleware if local.
      // For now, we return all tools.
      const tools = this.registry.getAllTools().map(t => ({
        name: t.name,
        description: t.description,
        inputSchema: zodToJsonSchema(t.input) as any,
      }));
      return { tools };
    });

    server.setRequestHandler(CallToolRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;
      const tool = this.registry.getAllTools().find(t => t.name === name);
      
      if (!tool) {
        throw new Error(`Tool not found: ${name}`);
      }

      // Mock identity for now, normally extracted from request context (HTTP header via proxy)
      const identity = { userId: 'system', role: 'admin', claims: {} };
      const domain = this.registry.getAll().find(d => d.tools.includes(tool))!.manifest;

      const ctx = {
        identity,
        domain,
        hooks: this.hooks,
        logger: console, // Temporary logger
        services: this.servicesMap,
        tools: {
          call: async (toolName: string, input: any) => {
            // Internal call logic
            return null as any;
          }
        },
        resources: {
          get: async (uri: string) => {
            return null as any;
          }
        }
      };

      let input = await this.hooks.applyFilters('filter:tool:input', args, ctx);
      
      // Validation
      const parsedInput = tool.input.parse(input);

      await this.hooks.doAction('before:tool:call', { toolName: name, input: parsedInput, identity, domain });

      if (tool.auth) await authMiddleware(tool, ctx);
      if (tool.approval) await approvalMiddleware(tool, parsedInput, ctx);

      const startTime = Date.now();
      let output = await tool.handler(parsedInput, ctx);
      const duration = Date.now() - startTime;

      output = await this.hooks.applyFilters('filter:tool:output', output, ctx);

      await this.hooks.doAction('after:tool:call', { toolName: name, input: parsedInput, output, identity, duration });

      if (tool.audit) await auditMiddleware(tool, parsedInput, output, ctx);

      return output as any;
    });
  }

  private registerResources(server: Server) {
    server.setRequestHandler(ListResourcesRequestSchema, async () => {
      const resources = this.registry.getAllResources().map(r => ({
        uri: r.uri,
        name: r.name,
        description: r.description,
        mimeType: r.mimeType,
      }));
      return { resources };
    });

    server.setRequestHandler(ReadResourceRequestSchema, async (request) => {
      const { uri } = request.params;
      const resource = this.registry.getAllResources().find(r => r.uri === uri); // Naive matching for now

      if (!resource) {
        throw new Error(`Resource not found: ${uri}`);
      }

      const identity = { userId: 'system', role: 'admin', claims: {} };
      const domain = this.registry.getAll().find(d => d.resources.includes(resource))!.manifest;
      
      const ctx = {
        identity,
        domain,
        hooks: this.hooks,
        logger: console,
        services: this.servicesMap,
      };

      await this.hooks.doAction('before:resource:read', { uri, identity });
      
      let data = await resource.handler({} as any, ctx); // Naive param extraction
      data = await this.hooks.applyFilters('filter:resource:data', data, ctx);

      await this.hooks.doAction('after:resource:read', { uri, data, identity });

      return {
        contents: [{ uri, mimeType: resource.mimeType, text: data }]
      };
    });
  }

  private registerPrompts(server: Server) {
    server.setRequestHandler(ListPromptsRequestSchema, async () => {
      const prompts = this.registry.getAllPrompts().map(p => ({
        name: p.name,
        description: p.description,
        arguments: [] // Should extract from Zod schema
      }));
      return { prompts };
    });

    server.setRequestHandler(GetPromptRequestSchema, async (request) => {
      const { name, arguments: args } = request.params;
      const prompt = this.registry.getAllPrompts().find(p => p.name === name);

      if (!prompt) {
        throw new Error(`Prompt not found: ${name}`);
      }

      const identity = { userId: 'system', role: 'admin', claims: {} };
      const domain = this.registry.getAll().find(d => d.prompts.includes(prompt))!.manifest;
      
      const ctx = {
        identity,
        domain,
        hooks: this.hooks,
        logger: console,
        services: this.servicesMap,
      };

      const parsedArgs = prompt.arguments.parse(args);
      const result = await prompt.handler(parsedArgs, ctx);

      return {
        description: prompt.description,
        messages: result.messages as any
      };
    });
  }
}
