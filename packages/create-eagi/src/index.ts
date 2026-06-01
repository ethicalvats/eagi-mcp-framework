#!/usr/bin/env node
import { spawnSync } from 'node:child_process';
import { Command } from 'commander';

const program = new Command('create-eagi')
  .argument('<project-name>', 'Name of your new project')
  .action((name) => {
    // A real bootstrapper would clone a template or construct it here.
    // For now, we simulate by invoking the `@eagi/cli`'s `init` command
    // assuming the CLI is globally installed, or we just write it directly.
    console.log(`Bootstrapping project: ${name}`);
    
    spawnSync('npx', ['-y', '@eagi/cli@latest', 'init', name], { stdio: 'inherit' });
  });

program.parse();
