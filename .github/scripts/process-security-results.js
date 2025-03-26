const fs = require('fs');
const { sarifToMarkdown } = require('@security-alert/sarif-to-markdown');

function run() {
  try {
    // Read SARIF files
    const codeSarif = JSON.parse(fs.readFileSync('code-reports/snyk-code-report.sarif', 'utf8'));
    const containerSarif = JSON.parse(fs.readFileSync('container-reports/snyk-container-report.sarif', 'utf8'));

    // Helper function to check if SARIF has results
    const hasResults = (sarif) => {
      console.log('Checking SARIF structure:', {
        hasRuns: !!sarif.runs,
        hasFirstRun: !!(sarif.runs && sarif.runs[0]),
        hasResults: !!(sarif.runs && sarif.runs[0] && sarif.runs[0].results),
        resultsLength: sarif.runs && sarif.runs[0] && sarif.runs[0].results ? sarif.runs[0].results.length : 'N/A',
        rulesLength: sarif.runs && sarif.runs[0] && sarif.runs[0].tool && sarif.runs[0].tool.driver && sarif.runs[0].tool.driver.rules ? sarif.runs[0].tool.driver.rules.length : 'N/A'
      });
      
      // Check for either results or rules
      return (sarif.runs && sarif.runs[0] && (
        (sarif.runs[0].results && sarif.runs[0].results.length > 0) ||
        (sarif.runs[0].tool && sarif.runs[0].tool.driver && sarif.runs[0].tool.driver.rules && sarif.runs[0].tool.driver.rules.length > 0)
      ));
    };

    console.log('Code SARIF:', JSON.stringify(codeSarif, null, 2));
    console.log('Container SARIF:', JSON.stringify(containerSarif, null, 2));

    // Convert SARIF to markdown or show "no issues found" message
    const codeResults = hasResults(codeSarif) 
      ? sarifToMarkdown({
          title: "Code Scan Results",
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })(codeSarif)
      : [{ body: "✅ No vulnerabilities found in code scan.", hasMessages: false, shouldFail: false }];

    const containerResults = hasResults(containerSarif)
      ? sarifToMarkdown({
          title: "Container Scan Results", 
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })(containerSarif)
      : [{ body: "✅ No vulnerabilities found in container scan.", hasMessages: false, shouldFail: false }];

    console.log('Code Results:', JSON.stringify(codeResults, null, 2));
    console.log('Container Results:', JSON.stringify(containerResults, null, 2));

    // Build comment
    const timestamp = new Date().toISOString();
    const retVal = {
      title: "🔒 Security Scan Results",
      body: `Last updated: ${timestamp}

## 📝 Code Scan
${codeResults[0].body}
${codeResults.some(r => r.shouldFail) ? '⚠️ High or Critical vulnerabilities found!' : ''}

## 🐳 Container Scan
${containerResults[0].body}
${containerResults.some(r => r.shouldFail) ? '⚠️ High or Critical vulnerabilities found!' : ''}`
    };

    console.log('Final return value:', JSON.stringify(retVal, null, 2));
    return retVal;
  } catch (error) {
    console.error('Error processing security results:', error);
    throw error;
  }
}

module.exports = { run }; 