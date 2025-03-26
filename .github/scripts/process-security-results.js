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
    const codeResult = hasResults(codeSarif) 
      ? sarifToMarkdown({
          title: "Code Scan Results",
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })(codeSarif)
      : { body: "✅ No vulnerabilities found in code scan.", hasMessages: false, shouldFail: false };

    const containerResult = hasResults(containerSarif)
      ? sarifToMarkdown({
          title: "Container Scan Results", 
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })(containerSarif)
      : { body: "✅ No vulnerabilities found in container scan.", hasMessages: false, shouldFail: false };

    // Build comment
    const timestamp = new Date().toISOString();
    const retVal = {
      title: "🔒 Security Scan Results",
      body: `Last updated: ${timestamp}

## 📝 Code Scan
${codeResult.body}
${codeResult.shouldFail ? '⚠️ High or Critical vulnerabilities found!' : ''}

## 🐳 Container Scan
${containerResult.body}
${containerResult.shouldFail ? '⚠️ High or Critical vulnerabilities found!' : ''}`
    };

    return retVal;
  } catch (error) {
    console.error('Error processing security results:', error);
    throw error;
  }
}

module.exports = { run }; 