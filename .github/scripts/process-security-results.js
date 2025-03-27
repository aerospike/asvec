const fs = require('fs');
const { sarifToMarkdown } = require('@security-alert/sarif-to-markdown');

function run() {
  try {
    // Read SARIF files
    const codeSarif = JSON.parse(fs.readFileSync('code-reports/snyk-code-report.sarif', 'utf8'));
    const containerSarif = JSON.parse(fs.readFileSync('container-reports/snyk-container-report.sarif', 'utf8'));

    // Helper function to check if SARIF has results
    const hasResults = (sarif) => {
      if (!sarif?.runs?.length) {
        console.log('No runs found in SARIF file');
        return false;
      }

      // Thisis all just for logging.
      console.log('SARIF overview:');
      const summary = sarif.runs.map((run, index) => ({
        run: index,
        tool: run?.tool?.driver?.name ?? 'Unknown',
        results: run?.results?.length ?? 0,
        rules: run?.tool?.driver?.rules?.length ?? 0,
        severity: run?.results?.map(r => r.level).filter(Boolean) ?? []
      }));
      console.table(summary);

      // Check all runs for either results or rules
      return sarif.runs.some(run => 
        run?.results?.length > 0 || 
        run?.tool?.driver?.rules?.length > 0
      );
    };

    // Convert SARIF to markdown or show "no issues found" message
    const codeResults = hasResults(codeSarif) 
      ? sarifToMarkdown({
          title: "Code Scan Results",
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })(codeSarif)
      : [{ 
          body: "### Code Scan Summary\n✅ No security vulnerabilities found\n\n" + 
                `_Analyzed with ${codeSarif.runs?.map(r => r?.tool?.driver?.name).filter(Boolean).join(", ")}_`,
          hasMessages: false, 
          shouldFail: false 
        }];

    const containerResults = hasResults(containerSarif)
      ? sarifToMarkdown({
          title: "Container Scan Results", 
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })(containerSarif)
      : [{ 
          body: "### Container Scan Summary\n✅ No security vulnerabilities found\n\n" + 
                `_Analyzed with ${containerSarif.runs?.map(run => run?.tool?.driver?.name).filter(Boolean).join(", ")}_`,
          hasMessages: false, 
          shouldFail: false 
        }];

    // Build comment
    const timestamp = new Date().toISOString();
    const retVal = {
      title: "🔒 Security Scan Results",
      body: `Last updated: ${timestamp}

## 📝 Code Scan
${codeResults.map(result => result?.body ?? '').join('\n')}
${codeResults.some(r => r?.shouldFail) ? '⚠️ High or Critical vulnerabilities found!' : ''}

## 🐳 Container Scan
${containerResults.map(result => result?.body ?? '').join('\n')}
${containerResults.some(r => r?.shouldFail) ? '⚠️ High or Critical vulnerabilities found!' : ''}`
    };

    console.log('Final return value:', JSON.stringify(retVal, null, 2));
    return retVal;
  } catch (error) {
    console.error('Error processing security results:', error);
    throw error;
  }
}

module.exports = { run }; 