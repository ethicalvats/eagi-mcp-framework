#!/usr/bin/env node
import { Command } from 'commander';
import { initCommand } from './commands/init.js';
import { addCommand } from './commands/add.js';
import { devCommand } from './commands/dev.js';
import { serveCommand, serveDomainCommand } from './commands/serve.js';

const program = new Command();

program
  .name('eagi')
  .description('Enterprise AGI - Framework CLI')
  .version('0.1.0');

program.addCommand(initCommand);
program.addCommand(addCommand);
program.addCommand(devCommand);
program.addCommand(serveCommand);
program.addCommand(serveDomainCommand);

program.parse();
