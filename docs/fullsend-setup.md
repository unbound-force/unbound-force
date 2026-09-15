# Fullsend Setup and Troubleshooting Guide

This documents how to provision fullsend end-to-end on a GCP
project, including inference, GitHub configuration, agent
enablement, and IAM. It also covers common failure modes and
how to diagnose them.

For the authoritative upstream docs, see:
- [Getting Inference](https://github.com/fullsend-ai/fullsend/blob/main/docs/guides/getting-started/getting-inference.md)
- [Configuring GitHub](https://github.com/fullsend-ai/fullsend/blob/main/docs/guides/getting-started/configuring-github.md)
- [Operations](https://github.com/fullsend-ai/fullsend/blob/main/docs/guides/getting-started/operations.md)
- [Choose a Runtime](https://github.com/fullsend-ai/fullsend/blob/main/docs/guides/getting-started/choosing-a-runtime.md)

---

## End-to-End Setup

### Prerequisites

- `gcloud` CLI installed and authenticated
- `fullsend` CLI installed ([releases](https://github.com/fullsend-ai/fullsend/releases))
- `gh` CLI installed and authenticated
- A GCP project with billing enabled
- Admin access to the target GitHub org/repo

### Step 1 — Enable Required GCP APIs

```bash
gcloud services enable \
  iam.googleapis.com \
  cloudresourcemanager.googleapis.com \
  aiplatform.googleapis.com \
  --project="$GCP_PROJECT"
```

### Step 2 — Enable Models in Model Garden

Third-party publisher models (Anthropic Claude, Meta Llama,
Mistral) on Vertex AI must be enabled manually per-project.
Google exposes no API for this.

- Go to the [Vertex AI Model Garden](https://console.cloud.google.com/vertex-ai/model-garden)
  in the GCP console
- Enable the required models (at minimum: Claude Opus and
  Sonnet for the default Claude Code runtime)
- If your organization manages GCP projects centrally, contact
  the cloud platform team to request enablement via your
  internal ticketing process

> **Important:** Skipping this step produces the same 403
> `PERMISSION_DENIED` error as a missing IAM binding. See the
> troubleshooting section below for how to tell them apart.

### Step 3 — Provision Inference (WIF)

Use the fullsend CLI to create the WIF pool, OIDC provider,
and IAM bindings:

```bash
# Org-wide (covers all repos in the org)
fullsend inference provision ORG \
  --project "$GCP_PROJECT"

# Or per-repo (scoped to a single repository)
fullsend inference provision ORG/REPO \
  --project "$GCP_PROJECT"
```

This is idempotent and safe to re-run. The output includes the
WIF provider URL needed for the next step.

> **Do not set up WIF manually** unless you have a specific
> reason. The CLI handles pool creation, provider configuration,
> IAM binding scope, and audience settings correctly. Manual
> setup is error-prone — see the troubleshooting section for
> common mistakes.

Verify the provisioning:

```bash
fullsend inference status ORG \
  --project "$GCP_PROJECT"
```

### Step 4 — Install GitHub Apps

Install the agent apps to your organization and grant them
access to the target repository:

| Role | Installation URL |
|------|-----------------|
| triage | https://github.com/apps/fullsend-ai-triage/installations/new |
| coder | https://github.com/apps/fullsend-ai-coder/installations/new |
| review | https://github.com/apps/fullsend-ai-review/installations/new |
| retro | https://github.com/apps/fullsend-ai-retro/installations/new |
| prioritize | https://github.com/apps/fullsend-ai-prioritize/installations/new |

Install only the apps for the agents you want to enable. You
must also pass `--agents` in the next step to match.

### Step 5 — Configure GitHub

```bash
fullsend github setup ORG/REPO \
  --inference-project "$GCP_PROJECT" \
  --inference-wif-provider "$WIF_PROVIDER_URL"
```

This creates the workflow file, secrets, and variables on the
repository. It prompts for the runtime (default: `claude`).

To enable only specific agents:

```bash
fullsend github setup ORG/REPO \
  --inference-project "$GCP_PROJECT" \
  --inference-wif-provider "$WIF_PROVIDER_URL" \
  --agents triage,review
```

### Step 6 — Verify

Open a new issue on the repository, or comment `/fs-triage`
on an existing issue. Check the Actions tab — the
`fullsend-ai-triage` bot should post a comment within a few
minutes.

### Required Secrets and Variables

After setup, the repository should have:

| Name | Type | Purpose |
|------|------|---------|
| `FULLSEND_GCP_WIF_PROVIDER` | Secret | WIF provider resource name |
| `FULLSEND_GCP_PROJECT_ID` | Secret | GCP project ID for Vertex AI |
| `FULLSEND_MINT_URL` | Variable | Token minting service URL |
| `FULLSEND_GCP_REGION` | Variable | GCP region for Vertex AI |
| `FULLSEND_PROJECT_NUMBER` | Variable | GitHub Projects V2 number (Prioritize agent) |
| `FULLSEND_REVIEW_CLIENT_ID` | Variable | OAuth client ID for the review agent app |

Verify with:

```bash
gh variable list --repo ORG/REPO
gh secret list --repo ORG/REPO
```

---

## IAM Requirements

The WIF principal needs `roles/aiplatform.user` on the GCP
project. The `fullsend inference provision` command grants
this automatically.

If managing IAM manually, note the scoping difference:

| Scope | Principal suffix | Use case |
|-------|-----------------|----------|
| Per-repo | `attribute.repository/ORG/REPO` | Single repo, tighter trust boundary |
| Org-wide | `attribute.repository_owner/ORG` | Multiple repos share one GCP project |

Per-repo scoping means only that specific repository can
authenticate. If you add more repos later, either re-run
`fullsend inference provision` for each one or use the
org-wide scope.

### Managed GCP Environments

If your GCP project is managed by an IT automation that runs
periodic IAM policy reconciliation, custom bindings may be
removed silently. To prevent this:

- Request that the WIF principal bindings be added to the
  managed IAM baseline
- Monitor for sudden auth failures after previously working
  setups

---

## Troubleshooting

### WIF Authentication Failure — `invalid_target`

```
failed to generate Google Cloud federated token:
{"error":"invalid_target","error_description":"The target
service indicated by the \"audience\" parameters is invalid.
This might either be because the pool or provider is disabled
or deleted or because it doesn't exist."}
```

**Causes:**
1. The `FULLSEND_GCP_WIF_PROVIDER` secret doesn't match an
   existing pool/provider (typo, wrong project number, or
   set before the pool was created)
2. The pool or provider was deleted or disabled

**Diagnosis:**

```bash
# Verify pool exists and is ACTIVE
gcloud iam workload-identity-pools list \
  --location=global --project=PROJECT_ID

# Verify provider exists and is ACTIVE
gcloud iam workload-identity-pools providers list \
  --location=global \
  --workload-identity-pool=POOL_NAME \
  --project=PROJECT_ID

# Check with fullsend CLI
fullsend inference status ORG \
  --project PROJECT_ID
```

**Fix:** Re-run `fullsend inference provision` to recreate,
then update the secret via `fullsend github setup` or
`fullsend github set`.

### Inference Failure — 403 `PERMISSION_DENIED`

```
Permission 'aiplatform.endpoints.predict' denied on resource
'//aiplatform.googleapis.com/projects/.../publishers/anthropic/
models/...' (or it may not exist)
```

This error has **two distinct causes** that produce the same
message:

| Cause | The `(or it may not exist)` hint | How to confirm |
|-------|----------------------------------|----------------|
| **Missing IAM binding** | The permission is denied | Check IAM bindings (see below) |
| **Model not enabled** | The model endpoint doesn't exist | Check Model Garden in GCP console |

**Diagnosis — IAM:**

```bash
gcloud projects get-iam-policy PROJECT_ID \
  --flatten="bindings[].members" \
  --filter="bindings.members:POOL_NAME" \
  --format="table(bindings.role, bindings.members)"
```

Check that `roles/aiplatform.user` is bound to the correct
principal scope (`repository_owner` for org-wide, or
`repository` matching the repo that runs the workflow).

**Diagnosis — Model Garden:**

Go to the GCP console > Vertex AI > Model Garden and check
whether the required publisher models are enabled on the
project. There is no `gcloud` command to check this.

**Fix:**
- For IAM: re-run `fullsend inference provision`
- For Model Garden: enable models manually (see Step 2 above)

### Missing `FULLSEND_PROJECT_NUMBER` Variable

The Prioritize agent uses `FULLSEND_PROJECT_NUMBER` to access
GitHub Projects V2 for RICE scoring. If this variable is not
set, the workflow input `project_number` resolves to an empty
string and the Prioritize agent fails silently.

```bash
gh variable set FULLSEND_PROJECT_NUMBER \
  --repo ORG/REPO \
  --body "PROJECT_NUMBER"
```

The project number is the GitHub Projects V2 project number
(visible in the project URL), not the GCP project number.

### Workflows Fail After Previously Working

If agent workflows start failing suddenly after a period of
working correctly:

1. **Check if IAM bindings were removed** — IT automation may
   reconcile IAM policies and remove custom bindings
2. **Check if the WIF pool/provider was disabled** — run
   `fullsend inference status`
3. **Check if secrets/variables were modified** — review the
   repository audit log

---

## Runtime Selection

The default runtime is `claude` (Claude Code on Vertex AI).
Alternative runtimes (`pi`, `codex`) are experimental. See
[Choose a Runtime](https://github.com/fullsend-ai/fullsend/blob/main/docs/guides/getting-started/choosing-a-runtime.md)
for details.

To change the runtime after setup:

```bash
# Via CLI
fullsend github setup ORG/REPO --runtime pi

# Or edit .fullsend/config.yaml directly
```

Per-agent runtime overrides are also supported:

```bash
fullsend agent set code --runtime claude
fullsend agent set triage --runtime pi
```
