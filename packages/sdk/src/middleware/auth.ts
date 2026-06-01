import type { ToolDefinition, ToolContext } from '../domain/types';

export class AuthError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'AuthError';
  }
}

export async function authMiddleware(tool: ToolDefinition, ctx: ToolContext): Promise<void> {
  if (!tool.auth || !tool.auth.roles) return;
  
  const userRole = ctx.identity.role;
  const domainRoles = ctx.domain.roles || {};
  
  const allowedRoles = new Set(tool.auth.roles);
  
  // Basic RBAC checking
  if (allowedRoles.has(userRole)) return;

  // Check if user role includes any allowed role
  const userRoleDef = domainRoles[userRole];
  if (userRoleDef && userRoleDef.includes) {
    for (const included of userRoleDef.includes) {
      if (allowedRoles.has(included)) return;
    }
  }

  throw new AuthError(`User role '${userRole}' is not authorized to execute tool '${tool.name}'`);
}
