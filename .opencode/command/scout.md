---
description: Invoke Pinkman OSS Scout to discover, trend-scan, audit, or report on open source projects.
---
<!-- scaffolded by uf vdev -->

# Command: /scout

## User Input

```text
$ARGUMENTS
```

## Description

Invoke Pinkman to scout open source projects. Pinkman
discovers projects by domain keyword, classifies
licenses against the OSI-approved list, lists direct
dependencies with overlap detection, tracks industry
trends, audits existing dependency health, and generates
structured adoption recommendation reports.

## Modes

| Mode     | Syntax                           | Description                                |
|----------|----------------------------------|--------------------------------------------|
| discover | `/scout <keyword>`               | Discover OSI-approved projects by keyword  |
| trend    | `/scout --trend <category>`      | Find trending projects in a category       |
| audit    | `/scout --audit [manifest-path]` | Audit dependencies from a manifest file    |
| report   | `/scout --report <project-url>`  | Generate recommendation report for project |

**Default mode**: `discover` (when no flag is provided).

## Execution

1. **Parse the mode** from `$ARGUMENTS`:
   - If `$ARGUMENTS` starts with `--trend `: extract the
     category after `--trend ` and set mode to `trend`.
   - If `$ARGUMENTS` starts with `--audit`: extract the
     optional manifest path (default: `go.mod`) and set
     mode to `audit`.
   - If `$ARGUMENTS` starts with `--report `: extract
     the project URL after `--report ` and set mode to
     `report`.
   - Otherwise: treat the entire `$ARGUMENTS` as the
     domain keyword and set mode to `discover`.

2. **Delegate to the Pinkman agent** using the Task tool
   with `subagent_type: "pinkman"`.

3. **Construct the prompt** for the agent:

   For **discover** mode:
   > "Scout open source projects for the domain:
   > <keyword>. Use Discover Mode. Search GitHub for
   > relevant repositories, classify licenses against
   > the OSI-approved list, list direct dependencies,
   > detect shared dependency overlaps, and present
   > results in the standard output format. Save the
   > report to .uf/pinkman/reports/."

   For **trend** mode:
   > "Find trending open source projects in the
   > category: <category>. Use Trend Mode. Search for
   > repositories with high recent activity, compute
   > trend indicators (star growth, release velocity,
   > contributor activity), classify licenses, and rank
   > by trend strength. Save the report to
   > .uf/pinkman/reports/."

   For **audit** mode:
   > "Audit the dependencies in <manifest-path>. Use
   > Audit Mode. Read the manifest file, check each
   > dependency for available updates, detect license
   > changes between versions, assess maintenance risk,
   > and present results in the audit table format. Save
   > the report to .uf/pinkman/reports/."

   For **report** mode:
   > "Generate an adoption recommendation report for
   > <project-url>. Use Report Mode. Fetch comprehensive
   > project metadata, analyze license, community health,
   > trends, maintenance, dependencies, and produce a
   > structured recommendation with a verdict (adopt /
   > evaluate / defer / avoid). Save the report to
   > .uf/pinkman/reports/."

4. **Return the result** from the Pinkman agent to the
   user.
