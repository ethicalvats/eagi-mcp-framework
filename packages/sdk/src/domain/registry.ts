import type { LoadedDomain } from './loader';
import type { ToolDefinition, ResourceDefinition, PromptDefinition, ServiceDefinition } from './types';

export class DomainRegistry {
  private domains: Map<string, LoadedDomain> = new Map();

  register(domain: LoadedDomain) {
    this.domains.set(domain.manifest.name, domain);
  }

  get(name: string): LoadedDomain | undefined {
    return this.domains.get(name);
  }

  getAll(): LoadedDomain[] {
    return Array.from(this.domains.values());
  }

  getAllTools(): ToolDefinition[] {
    return this.getAll().flatMap(d => d.tools);
  }

  getAllResources(): ResourceDefinition[] {
    return this.getAll().flatMap(d => d.resources);
  }

  getAllPrompts(): PromptDefinition[] {
    return this.getAll().flatMap(d => d.prompts);
  }
}
