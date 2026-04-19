import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const servicesDir = path.join(__dirname, 'src/api/services');
const files = fs.readdirSync(servicesDir).filter(f => f.endsWith('.ts'));

let exportsCode = '\n// Auto-generated exports from services\n';

for (const file of files) {
    const content = fs.readFileSync(path.join(servicesDir, file), 'utf8');
    const serviceNameMatch = content.match(/export const (\w+) = {/);
    if (serviceNameMatch) {
        const serviceName = serviceNameMatch[1];
        // Now find all keys inside the service object
        // Match lines like `  fetchAnalytics: (` or `  getAutoCategorizeStatus: ()`
        const methodMatches = [...content.matchAll(/^\s+([a-zA-Z0-9_]+):\s*(?:\([^)]*\)|[a-zA-Z0-9_]+|async\s*\()/gm)];
        if (methodMatches.length > 0) {
            const methods = methodMatches.map(m => m[1]);
            exportsCode += `import { ${serviceName} } from './services/${file.replace('.ts', '')}';\n`;
            exportsCode += `export const { ${methods.join(', ')} } = ${serviceName};\n`;
        }
    } else {
        // try to find other forms, like just "export const XXX = "
    }
}

fs.appendFileSync(path.join(__dirname, 'src/api/client.ts'), exportsCode);
console.log('Done generating exports');
