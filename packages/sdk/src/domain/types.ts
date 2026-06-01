import type { ZodType, z } from 'zod';
import type { Identity, DomainManifest, ToolResult, HookContextMap, FilterDataMap } from '../hooks/types';
export type { Identity, DomainManifest, ToolResult };
import type { HookEngine } from '../hooks/engine';

export interface ToolContext {
  identity: Identity;
  domain: DomainManifest;
  hooks: HookEngine;
  logger: any; // Will be AuditLogger
  services: Record<string, any>; // Injected domain services
  tools: {
    call: <T = unknown>(toolName: string, input: Record<string, unknown>) => Promise<T>;
  };
  resources: {
    get: <T = unknown>(uri: string) => Promise<T>;
  };
}

export interface ResourceContext {
  identity: Identity;
  domain: DomainManifest;
  hooks: HookEngine;
  logger: any;
  services: Record<string, any>;
}

export interface ToolDefinition<TInput extends ZodType = ZodType> {
  name: string;
  description: string;
  input: TInput;
  auth?: { roles: string[] };
  audit?: boolean | { redactFields?: string[] };
  approval?: {
    required: boolean;
    message: string | ((input: z.infer<TInput>) => string);
    timeout?: string;
  };
  handler: (input: z.infer<TInput>, ctx: ToolContext) => Promise<ToolResult>;
}

type ExtractParams<T extends string> =
  T extends `${string}{${infer Param}}${infer Rest}`
    ? Param | ExtractParams<Rest>
    : never;

export interface ResourceDefinition<TUri extends string = string> {
  uri: TUri;
  name: string;
  description: string;
  mimeType: string;
  auth?: { roles: string[] };
  handler: (
    params: Record<ExtractParams<TUri>, string>,
    ctx: ResourceContext
  ) => Promise<string>;
}

export interface PromptMessage {
  role: 'user' | 'assistant';
  content: { type: 'text'; text: string } | { type: 'image'; data: string; mimeType: string };
}

export interface PromptDefinition<TArgs extends ZodType = ZodType> {
  name: string;
  description: string;
  arguments: TArgs;
  handler: (args: z.infer<TArgs>, ctx: ResourceContext) => Promise<{ messages: PromptMessage[] }>;
}

export interface DomainConfig {
  env: Record<string, string | undefined>;
}

type ResolvedDeps<TDeps extends readonly string[], TServiceMap> = {
  [K in TDeps[number]]: K extends keyof TServiceMap ? TServiceMap[K] : never;
};

export interface ServiceDefinition<
  TName extends string = string,
  TDeps extends readonly string[] = readonly string[],
  TInstance = any
> {
  name: TName;
  deps?: TDeps;
  factory: (config: DomainConfig, deps: ResolvedDeps<TDeps, any>) => Promise<TInstance>;
  dispose?: (instance: TInstance) => Promise<void>;
}

export interface EagiConfig {
  name: string;
  version: string;
  gateway?: {
    port: number;
    auth?: {
      provider: string;
      issuer: string;
      audience: string;
    };
  };
  audit?: {
    hashChain?: boolean;
    output?: 'file' | 'stdout';
    redactInputs?: boolean;
  };
  hooks?: {
    [K in keyof HookContextMap]?: (ctx: HookContextMap[K]) => Promise<void> | void;
  } & {
    [K in keyof FilterDataMap]?: (data: FilterDataMap[K], ctx?: any) => Promise<FilterDataMap[K]> | FilterDataMap[K];
  };
  domains?: Record<string, { enabled: boolean }>;
}
