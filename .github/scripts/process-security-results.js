const fs = require('fs');
const sarifToMarkdown = require('@security-alert/sarif-to-markdown');

async function run() {
  try {
    const codeSarif = JSON.parse(fs.readFileSync('code-reports/snyk-code-report.sarif', 'utf8'));
    const containerSarif = JSON.parse(fs.readFileSync('container-reports/snyk-container-report.sarif', 'utf8'));
    
    const codeMarkdown = await sarifToMarkdown(codeSarif);
    const containerMarkdown = await sarifToMarkdown(containerSarif);
    
    const comment = `## 🔒 Security Scan Results

### 📝 Code Scan
${codeMarkdown}

### 🐳 Container Scan
${containerMarkdown}

<sub>Last updated: ${new Date().toISOString()}</sub>`;

    return comment;
  } catch (error) {
    console.error('Error processing security results:', error);
    throw error;
  }
}

module.exports = { run }; 