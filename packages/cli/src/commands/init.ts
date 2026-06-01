import { Command } from 'commander';
import { writeFileSync, mkdirSync } from 'node:fs';
import { join } from 'node:path';
import chalk from 'chalk';

export const initCommand = new Command('init')
  .description('Scaffold a new EAGI project')
  .argument('[name]', 'Project name', 'my-eagi-app')
  .action((name) => {
    console.log(chalk.blue(`Initializing EAGI project: ${name}...`));
    
    const targetDir = join(process.cwd(), name);
    mkdirSync(targetDir, { recursive: true });
    
    // Write package.json
    writeFileSync(join(targetDir, 'package.json'), JSON.stringify({
      name,
      version: '0.1.0',
      scripts: {
        "dev": "eagi dev",
        "serve": "eagi serve"
      },
      dependencies: {
        "@eagi/sdk": "^0.1.0",
        "zod": "^3.22.4"
      }
    }, null, 2));

    // Write eagi.config.ts
    writeFileSync(join(targetDir, 'eagi.config.ts'), `import { defineConfig } from '@eagi/sdk';

export default defineConfig({
  name: '${name}',
  version: '0.1.0',
  domains: {}
});
`);

    console.log(chalk.green(`✓ Project ${name} scaffolded.`));
    console.log(`\nNext steps:\n  cd ${name}\n  npm install\n  npx eagi add domain core\n  npm run dev`);
  });
