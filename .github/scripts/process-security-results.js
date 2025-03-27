const fs = require('fs');

function run() {
    // Read SARIF files
    const codeSarif = JSON.parse(fs.readFileSync('code-reports/snyk-code-report.sarif', 'utf8'));
    const containerSarif = JSON.parse(fs.readFileSync('container-reports/snyk-container-report.sarif', 'utf8'));

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
    };

    /**
     * Generates a Markdown summary from a SARIF file.
     * It iterates over all runs, and for each result it looks up the associated rule
     * from the tool section, then outputs a table with details including help markdown.
     *
     * @param {object} sarif - Parsed SARIF (json) object.
     * @returns {string} - Markdown summary.
     */
    const generateMarkdownSummary = (sarif) => {
        printResults(sarif);
        if (!sarif || !sarif.runs || sarif.runs.length === 0) {
            return "No runs found in the SARIF file.";
        }

        let md = "# SARIF Scan Summary\n\n";

        return sarif.runs.reduce((md, run, runIndex) => {
            const toolName = run?.tool?.driver?.name || "Unknown Tool";
            md += `## Run ${runIndex + 1} - Tool: **${toolName}**\n\n`;

            if (run.results && run.results.length > 0) {
                md += "| Severity | Rule ID | Message | File | Start Line |\n";
                md += "|---------|----------|---------|------|------------|\n";
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
                    const escapedHelp = helpMarkdown.replace(/\|/g, '\\|');
                    const severityColor = getSeverityColor(severity);

                    return `| <span style="color:${severityColor};font-weight:bold;">${severity}</span> | ${ruleId} | ${message} | ${location} | ${startLine} |\n` +
                           `| <td colspan="5"><details><summary>View Details</summary>\n\n${escapedHelp}\n\n</details></td> |\n`;
                }).join('');
            } else {
                md += "No issues found in this run.\n";
            }

            md += "\n";
            return md;
        }, "# SARIF Scan Summary\n\n");
    }

    // Build comment
    const timestamp = new Date().toISOString();
    const retVal = {
        title: "🔒 Security Scan Results",
        body: `Last updated: ${timestamp}

## 📝 Code Scan
${generateMarkdownSummary(codeSarif)}

## 🐳 Container Scan
${generateMarkdownSummary(containerSarif)}
`
    };

    console.log('Final return value:', JSON.stringify(retVal, null, 2));
    return retVal;
}

module.exports = { run }; 