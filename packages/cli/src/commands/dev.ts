import { Command } from 'commander';
import { EagiRunner } from '@eagi/sdk';
import { readdirSync, existsSync } from 'node:fs';
import { join } from 'node:path';
import chalk from 'chalk';

export const devCommand = new Command('dev')
  .description('Start the EAGI server in development mode')
  .action(async () => {
    console.log(chalk.yellow('Starting EAGI in DEV mode (hot-reload coming soon)...'));
    await runServer();
  });



async function runServer() {
  const cwd = process.cwd();
  
  // Load config
  const configPath = join(cwd, 'eagi.config.ts');
  let config = { name: 'eagi-server', version: '1.0.0' };
  if (existsSync(configPath)) {
    // In a real CLI, we would use something like jiti to load TS config
    config = require(join(cwd, 'eagi.config.js')) || config;
  }

  // Find domains
  const domainsDir = join(cwd, 'domains');
  let domainDirs: string[] = [];
  if (existsSync(domainsDir)) {
    domainDirs = readdirSync(domainsDir, { withFileTypes: true })
      .filter(dirent => dirent.isDirectory())
      .map(dirent => join(domainsDir, dirent.name));
  } else {
    // If running in a specific domain dir directly
    if (existsSync(join(cwd, 'domain.yaml'))) {
      domainDirs = [cwd];
    }
  }

  const runner = new EagiRunner(config as any);
  await runner.start(domainDirs);
  console.log(chalk.green('✓ EAGI Server started on stdio transport.'));
}
