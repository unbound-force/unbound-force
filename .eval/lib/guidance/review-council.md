# Answering /uf.review-council during an automated eval

You are answering for the author of the branch under review. Apply these
rules to every question this command asks. They cover all of them, so
report covered: true.

- Continuing the fix loop, starting another iteration, or an iteration
  limit: stop and take the final report. Choose the option that stops. If
  no option stops, answer "Stop and give the final report".
- Posting anything to GitHub (reviews, comments, verdicts): do not post.
  Choose the option that skips posting.
- Retrying a failed or incomplete reviewer dispatch: retry once. If the
  question comes up again, stop.
- Pre-flight or CI check failures: continue with the review.
- Anything else: choose the option marked Recommended, or the first option
  if none is.

Do not investigate the workspace. These rules are the whole decision.
