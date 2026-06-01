export interface Identity {
  userId: string;
  role: string;
  email?: string;
  claims: Record<string, unknown>;
}

export interface DomainManifest {
  name: string;
  version: string;
  description?: string;
  dependencies?: string[];
  roles?: Record<string, { description?: string; includes?: string[] }>;
  triggers?: any[];
}

export interface ToolResult {
  content: Array<{ type: 'text'; text: string } | { type: 'image'; data: string; mimeType: string }>;
  isError?: boolean;
}

export interface HookContextMap {
  'before:tool:call': { toolName: string; input: unknown; identity: Identity; domain: DomainManifest };
  'after:tool:call': { toolName: string; input: unknown; output: ToolResult; identity: Identity; duration: number };
  'on:tool:error': { toolName: string; input: unknown; error: Error; identity: Identity };
  'before:resource:read': { uri: string; identity: Identity };
  'after:resource:read': { uri: string; data: string; identity: Identity };
  'on:server:start': { config: any; domains: DomainManifest[] };
  'on:server:stop': { uptime: number; requestCount: number };
  'on:domain:load': { domain: DomainManifest };
}

export interface FilterDataMap {
  'filter:tool:input': unknown;
  'filter:tool:output': ToolResult;
  'filter:resource:data': string;
  'filter:tools:list': any[]; // Will be ToolDefinition[]
  'filter:audit:entry': any; // Will be AuditEntry
}

export type ActionHandler<K extends keyof HookContextMap> = (ctx: HookContextMap[K]) => Promise<void> | void;
export type FilterHandler<K extends keyof FilterDataMap> = (
  data: FilterDataMap[K],
  ctx?: any
) => Promise<FilterDataMap[K]> | FilterDataMap[K];
