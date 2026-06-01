import { Command } from 'commander';
import { writeFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';
import chalk from 'chalk';

export const addCommand = new Command('add')
  .description('Add a new domain module')
  .argument('<type>', 'Type to add (currently only "domain")')
  .argument('<name>', 'Name of the domain')
  .action((type, name) => {
    if (type !== 'domain') {
      console.error(chalk.red('Only "domain" is supported right now.'));
      process.exit(1);
    }

    console.log(chalk.blue(`Scaffolding domain: ${name}...`));
    
    const domainDir = join(process.cwd(), 'domains', name);
    mkdirSync(join(domainDir, 'tools'), { recursive: true });
    mkdirSync(join(domainDir, 'resources'), { recursive: true });
    mkdirSync(join(domainDir, 'prompts'), { recursive: true });
    mkdirSync(join(domainDir, 'services'), { recursive: true });
    mkdirSync(join(domainDir, 'hooks'), { recursive: true });

    // domain.yaml
    writeFileSync(join(domainDir, 'domain.yaml'), `name: ${name}\nversion: 1.0.0\ndescription: ${name} domain\n`);
    
    // hooks/index.ts
    writeFileSync(join(domainDir, 'hooks', 'index.ts'), `import { defineHooks } from '@eagi/sdk';\n\nexport default defineHooks({});\n`);

    console.log(chalk.green(`✓ Domain ${name} created.`));
  });
