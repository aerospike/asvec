const fs = require('fs');

function run() {
    // Read SARIF files
    const codeSarif = JSON.parse(fs.readFileSync('code-reports/snyk-code-report.sarif', 'utf8'));
    const containerSarif = JSON.parse(fs.readFileSync('container-reports/snyk-container-report.sarif', 'utf8'));
    /**
     * This seems to be ignored by github PR comments. Still keeping it here in the hopes that it will be used in the future.
     * @param {*} severity 
     * @returns 
     */
    const getSeverityColor = (severity) => {
        const colors = {
            critical: '#cc0000',
            high: '#ff4444',
            medium: '#ff8800',
            moderate: '#ff8800',
            low: '#ffcc00',
            undefined: '#6c757d'
        };
        return colors[severity?.toLowerCase()] ?? colors.undefined;
    };

    /**
     * Prints a summary of the SARIF file to the console.
     * @param {object} sarif - The SARIF object.
     * @returns {boolean} - True if there are results, false otherwise.
     */
    const printResults = (sarif) => {
        if (!sarif?.runs?.length) {
            console.log('No runs found in SARIF file');
            return false;
        }

        // This is all just for logging.
        console.log('SARIF overview:');
        const summary = sarif.runs.map((run, index) => ({
            run: index,
            tool: run?.tool?.driver?.name ?? 'Unknown',
            results: run?.results?.length ?? 0,
            rules: run?.tool?.driver?.rules?.length ?? 0,
            severity: run?.results?.map(r => r.level).filter(Boolean) ?? []
        }));
        console.table(summary);
    };

    const getSeverityEmoji = (severity) => {
        const emojis = {
            critical: '🔴 🚨',  // Red circle + siren for critical
            high: '🟠 ⚠️',     // Orange circle + warning for high
            medium: '🟡 ⚠️',    // Yellow circle + warning for medium
            moderate: '🟡 ⚠️',   // Yellow circle + warning for moderate (alias of medium)
            low: '🔵 ℹ️',      // Blue circle + info for low
            undefined: '⚪ ❓'   // White circle + question mark for undefined
        };
        return emojis[severity?.toLowerCase()] ?? emojis.undefined;
    };

    /**
     * Generates a Markdown summary from a SARIF file.
     * It iterates over all runs, and for each result it looks up the associated rule
     * from the tool section, then outputs a table with details including help markdown.
     *
     * @param {object} sarif - Parsed SARIF (json) object.
     * @returns {string} - Markdown summary.
     */
    const processSarif = (sarif) => {
        printResults(sarif);
        if (!sarif || !sarif.runs || sarif.runs.length === 0) {
            return "No runs found in the SARIF file.";
        }

        let md = "\n";

        return sarif.runs.reduce((md, run, runIndex) => {
            const toolName = run?.tool?.driver?.name || "Unknown Tool";
            md += `> **Run ${runIndex + 1} - Tool: ${toolName}**\n\n`;
            let errors = 0;
            if (run.results && run.results.length > 0) {
                // Sarif schema is overly flexible, so we need to handle some weird cases. This is working for snyk output. 
                // It will problaly need to be adapted for other tools.
                md += run.results.map(result => {
                    const ruleId = result.ruleId || "N/A";
                    const rule = run?.tool?.driver?.rules?.find(r => r.id === ruleId) || {};
                    const ruleDesc = rule.shortDescription?.text || "";
                    const severity = result.level || rule.defaultConfiguration?.level || "unknown";
                    const message = result.message?.text || ruleDesc || "No message";
                    const location = result.locations && result.locations.length > 0
                        ? result.locations[0].physicalLocation?.artifactLocation?.uri || "unknown"
                        : "unknown";
                    const startLine = result.locations && result.locations.length > 0
                        ? result.locations[0].physicalLocation?.region?.startLine || "N/A"
                        : "N/A";
                    const helpMarkdown = rule.help?.markdown || rule.help?.text || "";
                    const severityColor = getSeverityColor(severity);
                    const severityEmoji = getSeverityEmoji(severity);
    
                    return `<table>
<tr>
  <th>Severity</th>
  <th>Rule ID</th>
  <th>Message</th>
  <th>File</th>
  <th>Start Line</th>
</tr>
<tr>
  <td><span style="color:${severityColor};font-weight:bold;">${severityEmoji} ${severity}</span></td>
  <td>${ruleId}</td>
  <td>${message}</td>
  <td>${location}</td>
  <td>${startLine}</td>
</tr>
</table>

<details><summary>View Details</summary>

${helpMarkdown}

</details>

`;
                }).join('');
            } else {
                md += "> ✅ No security issues found in this run.\n";
            }

            md += "\n";
            return md;
        }, "\n\n");
    }

    // Build comment
    const timestamp = new Date().toISOString();
    const codeResults = processSarif(codeSarif);
    const containerResults = processSarif(containerSarif);
    const hasResults = codeResults.length > 0 || containerResults.length > 0;
    
    const retVal = {
        title: "🔒 Security Scan Results",
        body: `Last updated: ${timestamp}
    
**📝 Code Scan**
${codeResults}

**🐳 Container Scan**
${containerResults}

`
    };

    console.log('Final return value:', JSON.stringify(retVal, null, 2));
    return retVal;
}

module.exports = { run }; 