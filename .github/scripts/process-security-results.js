const fs = require('fs');
const { sarifToMarkdown } = require('@security-alert/sarif-to-markdown');

async function run() {
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
      ? await sarifToMarkdown(codeSarif, {
          title: "Code Scan Results",
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })
      : "✅ No vulnerabilities found in code scan.";

    const containerMarkdown = hasResults(containerSarif)
      ? await sarifToMarkdown(containerSarif, {
          title: "Container Scan Results", 
          severities: ["critical", "high", "medium", "low"],
          simpleMode: false,
          details: true,
          failOn: ["critical", "high"]
        })
      : "✅ No vulnerabilities found in container scan.";

    // Build comment
    const timestamp = new Date().toISOString();
    const comment = `# 🔒 Security Scan Results
Last updated: ${timestamp}

## 📝 Code Scan
${codeMarkdown}

## 🐳 Container Scan
${containerMarkdown}`;

    // Write comment to file for GitHub Actions to use
    fs.writeFileSync('security-comment.md', comment);
    console.log('Successfully processed security results');

  } catch (error) {
    console.error('Error processing security results:', error);
    process.exit(1);
  }
}

run(); 