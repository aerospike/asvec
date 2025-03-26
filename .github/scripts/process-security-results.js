const fs = require('fs');
const { sarifToMarkdown } = require('@security-alert/sarif-to-markdown');

function run() {
  try {
    // Read SARIF files
    const codeSarif = JSON.parse(fs.readFileSync('code-reports/snyk-code-report.sarif', 'utf8'));
    const containerSarif = JSON.parse(fs.readFileSync('container-reports/snyk-container-report.sarif', 'utf8'));

    // Helper function to check if SARIF has results
    const hasResults = (sarif) => {
      return sarif.runs && sarif.runs[0] && sarif.runs[0].results && sarif.runs[0].results.length > 0;
    };

    // Convert SARIF to markdown or show "no issues found" message
    const codeMarkdown = hasResults(codeSarif) 
      ?  sarifToMarkdown(codeSarif, {
          title: "Code Scan Results",
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })
      : "✅ No vulnerabilities found in code scan.";

    const containerMarkdown = hasResults(containerSarif)
      ?  sarifToMarkdown(containerSarif, {
          title: "Container Scan Results", 
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })
      : "✅ No vulnerabilities found in container scan.";

    // Build comment
    const timestamp = new Date().toISOString();
    return {
      title: "🔒 Security Scan Results",
      body: `Last updated: ${timestamp}

## 📝 Code Scan
${codeMarkdown}

## 🐳 Container Scan
${containerMarkdown}`
    };

  } catch (error) {
    console.error('Error processing security results:', error);
    throw error;
  }
}

module.exports = { run }; 