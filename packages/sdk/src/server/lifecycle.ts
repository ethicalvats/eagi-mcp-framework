import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { EagiServerBuilder } from './builder';
import { DomainRegistry, DomainLoader, sortServices } from '../domain';
import { HookEngine } from '../hooks';
import type { EagiConfig } from '../domain/types';

export class EagiRunner {
  private registry = new DomainRegistry();
  private hooks = new HookEngine();
  private servicesMap: Record<string, any> = {};

  constructor(private config: EagiConfig) {}

  async start(domainDirs: string[]) {
    // 1. Register global hooks from config
    if (this.config.hooks) {
      for (const [hookName, handler] of Object.entries(this.config.hooks)) {
        if (hookName.startsWith('filter:')) {
          this.hooks.addFilter(hookName as any, handler as any);
        } else {
          this.hooks.addAction(hookName as any, handler as any);
        }
      }
    }

    // 2. Load domains
    const loader = new DomainLoader();
    for (const dir of domainDirs) {
      const domain = await loader.load(dir);
      this.registry.register(domain);
      
      // Load domain-specific hooks
      if (domain.hooksModule && domain.hooksModule.default) {
        const domainHooks = domain.hooksModule.default;
        for (const [hookName, handler] of Object.entries(domainHooks)) {
          if (hookName.startsWith('filter:')) {
            this.hooks.addFilter(hookName as any, handler as any);
          } else {
            this.hooks.addAction(hookName as any, handler as any);
          }
        }
      }

      await this.hooks.doAction('on:domain:load', { domain: domain.manifest });
    }

    // 3. Initialize Services (Topological Sort)
    const allServices = this.registry.getAll().flatMap(d => d.services);
    const sortedServices = sortServices(allServices);

    for (const service of sortedServices) {
      const deps: Record<string, any> = {};
      if (service.deps) {
        for (const dep of service.deps) {
          deps[dep] = this.servicesMap[dep];
        }
      }
      // Initialize the service factory
      const instance = await service.factory({ env: process.env as any }, deps);
      this.servicesMap[service.name] = instance;
    }

    // 4. Build MCP Server
    const builder = new EagiServerBuilder(this.config, this.registry, this.hooks, this.servicesMap);
    const server = builder.build();

    // 5. Connect Transport
    const transport = new StdioServerTransport();
    await server.connect(transport);

    await this.hooks.doAction('on:server:start', { config: this.config, domains: this.registry.getAll().map(d => d.manifest) });
  }
}
