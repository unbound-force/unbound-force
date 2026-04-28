## ADDED Requirements

### Requirement: License Compatibility Tier

After classifying a project's license as OSI-approved
or not, Pinkman MUST assign a compatibility tier:

| Tier | Licenses |
|------|----------|
| `permissive` | MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, Unlicense, 0BSD, Zlib, BSL-1.0 |
| `weak-copyleft` | LGPL-2.1-only, LGPL-2.1-or-later, LGPL-3.0-only, LGPL-3.0-or-later, MPL-2.0, EPL-2.0, EUPL-1.2, Artistic-2.0 |
| `strong-copyleft` | GPL-2.0-only, GPL-2.0-or-later, GPL-3.0-only, GPL-3.0-or-later, AGPL-3.0-only, AGPL-3.0-or-later |
| `unknown` | Any license not in the above tiers |

#### Scenario: Permissive license classified correctly
- **GIVEN** a project with license `MIT`
- **WHEN** Pinkman runs License Classification
- **THEN** the compatibility tier MUST be `permissive`

#### Scenario: Strong-copyleft license classified correctly
- **GIVEN** a project with license `GPL-3.0-only`
- **WHEN** Pinkman runs License Classification
- **THEN** the compatibility tier MUST be
  `strong-copyleft`

#### Scenario: Weak-copyleft license classified correctly
- **GIVEN** a project with license `MPL-2.0`
- **WHEN** Pinkman runs License Classification
- **THEN** the compatibility tier MUST be
  `weak-copyleft`

#### Scenario: Unrecognized license gets unknown tier
- **GIVEN** a project with an OSI-approved license not
  in the tier table (e.g., a newly approved license)
- **WHEN** Pinkman runs License Classification
- **THEN** the compatibility tier MUST be `unknown`

### Requirement: Compatibility Verdict

Pinkman MUST produce a compatibility verdict for each
project based on its compatibility tier and the
reference license (Apache-2.0):

| Tier | Verdict |
|------|---------|
| `permissive` | `compatible` |
| `weak-copyleft` | `caution` |
| `strong-copyleft` | `incompatible` |
| `unknown` | `caution` |

For non-OSI licenses (`not_approved` verdict), the
compatibility verdict MUST be `incompatible`.

For non-standard licenses (`manual_review` verdict),
the compatibility verdict MUST be `caution`.

#### Scenario: Non-OSI license is incompatible
- **GIVEN** a project with a non-OSI-approved license
  (e.g., SSPL-1.0)
- **WHEN** Pinkman produces the compatibility verdict
- **THEN** the verdict MUST be `incompatible`
- **AND** the recommendation MUST be `avoid`

#### Scenario: Non-standard license gets caution
- **GIVEN** a project with a custom/non-standard
  license text
- **WHEN** Pinkman produces the compatibility verdict
- **THEN** the verdict MUST be `caution`
- **AND** the recommendation MUST NOT exceed `evaluate`

#### Scenario: Permissive license is compatible
- **GIVEN** a project with license `Apache-2.0`
- **WHEN** Pinkman produces the compatibility verdict
- **THEN** the verdict MUST be `compatible`

#### Scenario: GPL project is incompatible
- **GIVEN** a project with license `GPL-3.0-only`
- **WHEN** Pinkman produces the compatibility verdict
- **THEN** the verdict MUST be `incompatible`

#### Scenario: LGPL project gets caution
- **GIVEN** a project with license `LGPL-3.0-only`
- **WHEN** Pinkman produces the compatibility verdict
- **THEN** the verdict MUST be `caution`

### Requirement: Dual-License Compatibility

For dual-licensed projects (SPDX `OR` expression),
Pinkman MUST evaluate each license option
independently and use the most favorable (least
restrictive) compatibility tier.

Tier ordering from most to least favorable:
`permissive` > `weak-copyleft` > `strong-copyleft`
> `unknown`. The `unknown` tier is ranked least
favorable because an unclassified license carries
unbounded risk — at least `strong-copyleft` obligations
are well-understood.

#### Scenario: Dual-license with permissive option
- **GIVEN** a project with license `MIT OR GPL-3.0-only`
- **WHEN** Pinkman evaluates compatibility
- **THEN** the compatibility tier MUST be `permissive`
  (from MIT)
- **AND** the compatibility verdict MUST be `compatible`

#### Scenario: Dual-license both copyleft
- **GIVEN** a project with license `LGPL-3.0-only OR GPL-3.0-only`
- **WHEN** Pinkman evaluates compatibility
- **THEN** the compatibility tier MUST be
  `weak-copyleft` (from LGPL-3.0-only, more favorable)
- **AND** the compatibility verdict MUST be `caution`

#### Scenario: Dual-license with unknown option
- **GIVEN** a project with license
  `GPL-3.0-only OR FooBarLicense`
- **WHEN** Pinkman evaluates compatibility
- **THEN** the compatibility tier MUST be
  `strong-copyleft` (from GPL-3.0-only, more favorable
  than `unknown`)
- **AND** the compatibility verdict MUST be
  `incompatible`

### Non-Goal: SPDX `AND` Expressions

Conjunctive license expressions (SPDX `AND`, e.g.,
`Apache-2.0 AND GPL-3.0-only`) are out of scope for
this change. When Pinkman encounters an `AND`
expression, it MUST classify the compatibility tier
as `unknown` and produce a `caution` verdict. This is
a conservative default that requires human legal
review — `AND` means both licenses apply
simultaneously, and the interaction between
conjunctive obligations is too nuanced for automated
classification.

#### Scenario: AND expression gets caution default
- **GIVEN** a project with license
  `Apache-2.0 AND GPL-3.0-only`
- **WHEN** Pinkman evaluates compatibility
- **THEN** the compatibility tier MUST be `unknown`
- **AND** the compatibility verdict MUST be `caution`
- **AND** the recommendation MUST NOT exceed `evaluate`

### Non-Goal: SPDX `WITH` License Exceptions

License exceptions (SPDX `WITH`, e.g.,
`GPL-2.0-only WITH Classpath-exception-2.0`) are out
of scope for this change. When Pinkman encounters a
`WITH` expression, it MUST classify the compatibility
tier as `unknown` and produce a `caution` verdict.
License exceptions can materially change copyleft
obligations (e.g., the Classpath exception removes
linking requirements), but mapping individual
exceptions to tier adjustments is too complex for v1.
The `caution` verdict directs users to seek legal
review.

#### Scenario: WITH expression gets caution default
- **GIVEN** a project with license
  `GPL-2.0-only WITH Classpath-exception-2.0`
- **WHEN** Pinkman evaluates compatibility
- **THEN** the compatibility tier MUST be `unknown`
- **AND** the compatibility verdict MUST be `caution`
- **AND** the recommendation MUST NOT exceed `evaluate`

### Requirement: Compatibility-Gated Recommendation

The compatibility verdict MUST act as a hard gate on
the recommendation verdict:

| Compatibility | Maximum recommendation |
|---------------|-----------------------|
| `compatible` | `adopt` |
| `caution` | `evaluate` |
| `incompatible` | `avoid` |

#### Scenario: Healthy GPL project gets avoid
- **GIVEN** a project with license `GPL-3.0-only`,
  healthy maintenance, positive trend trajectory,
  and no dependency conflicts
- **WHEN** Pinkman assigns the recommendation verdict
- **THEN** the verdict MUST be `avoid`
- **AND** the reason MUST reference the license
  incompatibility with Apache-2.0

#### Scenario: Healthy LGPL project capped at evaluate
- **GIVEN** a project with license `LGPL-3.0-only`,
  healthy maintenance, positive trend trajectory,
  and no dependency conflicts
- **WHEN** Pinkman assigns the recommendation verdict
- **THEN** the verdict MUST NOT be `adopt`
- **AND** the verdict MUST be `evaluate` at most
- **AND** the reason MUST note the weak-copyleft
  `caution` status and recommend legal review

#### Scenario: Healthy permissive project can get adopt
- **GIVEN** a project with license `MIT`, healthy
  maintenance, positive trend trajectory, and no
  dependency conflicts
- **WHEN** Pinkman assigns the recommendation verdict
- **THEN** the verdict MAY be `adopt`
  (compatibility does not block it)

### Requirement: Compatibility in Output Formats

All Pinkman output formats MUST include the
compatibility tier and verdict for each project:

- **Discover/Trend Result List**: Add
  `- **Compatibility**: <tier> (<verdict>)` line
  after the License line.
- **Audit Result Table**: Add a `Compatibility`
  column after the `License Changed?` column.
- **Recommendation Report**: Add a
  `- **Compatibility**: <tier> (<verdict>)` line
  in the License Analysis section.

#### Scenario: Discover output shows compatibility
- **GIVEN** Pinkman discovers a project with license
  `GPL-3.0-only`
- **WHEN** the output is formatted
- **THEN** the project entry MUST include
  `- **Compatibility**: strong-copyleft (incompatible)`

### Requirement: Compatibility in Dewey Learnings

The structured prose in Dewey learnings (per the
Dewey Integration section in pinkman.md) MUST include
the compatibility verdict for each project.

#### Scenario: Dewey learning includes compatibility
- **GIVEN** Pinkman discovers 3 projects with verdicts
  `compatible`, `caution`, and `incompatible`
- **WHEN** the learning is stored via
  `dewey_store_learning`
- **THEN** the information string MUST include the
  compatibility verdict for each project (e.g.,
  "testify (MIT, permissive/compatible, adopt)")

## MODIFIED Requirements

### Requirement: License Classification (Spec 032)

Previously: License Classification produces an OSI
verdict (approved, not_approved, unknown,
manual_review, dual_approved) and stops.

Updated: After producing the OSI verdict, License
Classification MUST also assign a compatibility tier
and produce a compatibility verdict per the
requirements above. The OSI verdict is unchanged.

### Requirement: Recommendation Verdict (Spec 032)

Previously: `adopt` requires "OSI-approved license,
healthy maintenance, positive trend trajectory, no
dependency conflicts." `avoid` requires "License is
not OSI-approved, or critical supply chain risks."

Updated: `adopt` additionally requires `compatible`
compatibility verdict. `avoid` additionally includes
`incompatible` compatibility verdict as a trigger.
`evaluate` is the maximum for `caution` compatibility
verdict. All other criteria remain unchanged.

### Requirement: Fallback License List (Spec 032)

Previously: A flat comma-separated list of
OSI-approved SPDX identifiers.

Updated: Each license in the fallback list MUST
include its compatibility tier annotation in
parentheses (e.g., `MIT (permissive)`,
`GPL-3.0-only (strong-copyleft)`).

## REMOVED Requirements

None.
