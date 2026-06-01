import { readFileSync, existsSync, readdirSync } from 'node:fs';
import { join } from 'node:path';
import yaml from 'js-yaml';
import type { DomainManifest, ServiceDefinition, ToolDefinition, ResourceDefinition, PromptDefinition } from './types';

export interface LoadedDomain {
  manifest: DomainManifest;
  services: ServiceDefinition[];
  tools: ToolDefinition[];
  resources: ResourceDefinition[];
  prompts: PromptDefinition[];
  hooksModule?: any;
  dir: string;
}

export class DomainLoader {
  async load(domainDir: string): Promise<LoadedDomain> {
    const manifestPath = join(domainDir, 'domain.yaml');
    if (!existsSync(manifestPath)) {
      throw new Error(`Domain manifest not found at ${manifestPath}`);
    }

    const manifestContent = readFileSync(manifestPath, 'utf8');
    const manifest = yaml.load(manifestContent) as DomainManifest;

    const services = await this.loadDirectory<ServiceDefinition>(domainDir, 'services');
    const tools = await this.loadDirectory<ToolDefinition>(domainDir, 'tools');
    const resources = await this.loadDirectory<ResourceDefinition>(domainDir, 'resources');
    const prompts = await this.loadDirectory<PromptDefinition>(domainDir, 'prompts');
    
    let hooksModule;
    const hooksPath = join(domainDir, 'hooks', 'index.ts');
    const hooksPathJs = join(domainDir, 'hooks', 'index.js');
    if (existsSync(hooksPath) || existsSync(hooksPathJs)) {
      hooksModule = await import(join(domainDir, 'hooks', 'index'));
    }

    return {
      manifest,
      services,
      tools,
      resources,
      prompts,
      hooksModule,
      dir: domainDir
    };
  }

  private async loadDirectory<T>(domainDir: string, subDir: string): Promise<T[]> {
    const dirPath = join(domainDir, subDir);
    if (!existsSync(dirPath)) return [];

    const files = readdirSync(dirPath).filter(f => f.endsWith('.ts') || f.endsWith('.js'));
    const loaded: T[] = [];

    for (const file of files) {
      // Dynamic import
      const module = await import(join(dirPath, file));
      if (module.default) {
        loaded.push(module.default);
      }
    }

    return loaded;
  }
}

export function sortServices(services: ServiceDefinition[]): ServiceDefinition[] {
  const sorted: ServiceDefinition[] = [];
  const visited = new Set<string>();
  const visiting = new Set<string>();

  const serviceMap = new Map(services.map(s => [s.name, s]));

  function visit(name: string) {
    if (visited.has(name)) return;
    if (visiting.has(name)) {
      throw new Error(`Circular dependency detected involving service: ${name}`);
    }

    visiting.add(name);

    const service = serviceMap.get(name);
    if (service && service.deps) {
      for (const dep of service.deps) {
        if (!serviceMap.has(dep)) {
          throw new Error(`Service ${name} depends on unknown service ${dep}`);
        }
        visit(dep);
      }
    }

    visiting.delete(name);
    visited.add(name);
    if (service) sorted.push(service);
  }

  for (const service of services) {
    visit(service.name);
  }

  return sorted;
}
