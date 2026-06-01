import { Command } from 'commander';
import { EagiRunner } from '@eagi/sdk';
import { spawn } from 'node:child_process';
import { readdirSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import chalk from 'chalk';

export const serveCommand = new Command('serve')
  .description('Start the EAGI Gateway Control Plane (Production mode)')
  .action(() => {
    console.log(chalk.blue('Starting EAGI Gateway Control Plane...'));
    
    // In a real framework, this would execute a pre-compiled binary.
    // For this monorepo MVP, we assume the gateway is built.
    const gatewayPath = join(process.cwd(), 'gateway', 'eagi-gateway');
    
    // Run the gateway
    const child = spawn(gatewayPath, [], {
      cwd: process.cwd(),
      stdio: 'inherit',
      env: { ...process.env, DOMAIN_DIR: process.cwd() }
    });

    child.on('error', (err) => {
      // Fallback to go run if binary not found
      console.log(chalk.yellow('Compiled binary not found, attempting go run...'));
      const repoRoot = join(__dirname, '../../../../');
      const goChild = spawn('go', ['run', './cmd/eagi-gateway'], {
        cwd: join(repoRoot, 'gateway'),
        stdio: 'inherit',
        env: { ...process.env, DOMAIN_DIR: process.cwd() }
      });
      goChild.on('error', (e) => console.error(chalk.red("Failed to start: " + e.message)));
    });
  });

export const serveDomainCommand = new Command('serve-domain')
  .description('Internal: Start a specific Node.js domain server')
  .option('--domain <name>', 'Name of the domain to serve')
  .action(async (options) => {
    if (!options.domain) {
      console.error(chalk.red('--domain is required'));
      process.exit(1);
    }
    
    const cwd = process.cwd();
    const configPath = join(cwd, 'eagi.config.ts');
    let config = { name: 'eagi-server', version: '1.0.0' };
    
    if (existsSync(configPath)) {
      try { config = require(join(cwd, 'eagi.config.js')) || config; } catch(e) {}
    }

    const domainDirs = [join(cwd, 'domains', options.domain)];
    
    const runner = new EagiRunner(config as any);
    await runner.start(domainDirs);
  });
