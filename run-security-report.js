const processor = require('./.github/scripts/process-security-results.js');
            
async function updatePR() {
  // Get the comment content
  const result = await processor.run();
  
  // Find existing comment
  const { data: comments } = await github.rest.issues.listComments({
    owner: context.repo.owner,
    repo: context.repo.repo,
    issue_number: context.issue.number,
  });
  
  const botComment = comments.find(c => 
    c.user.type === 'Bot' && 
    c.body.includes(result.title)
  );
  
  const commentBody = `# ${result.title}\n${result.body}`;
  
  if (botComment) {
    // Update existing comment
    await github.rest.issues.updateComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      comment_id: botComment.id,
      body: commentBody
    });
  } else {
    // Create new comment
    await github.rest.issues.createComment({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: context.issue.number,
      body: commentBody
    });
  }
}

// Execute the async function
 updatePR().catch(error => {
    console.error('Failed to process security results:', error);
    process.exit(1);
  });