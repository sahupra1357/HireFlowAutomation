<!-- TEMPLATE for jobs/input/config/application-answers.md — /job setup copies this there if the file is missing. The live copy is gitignored. -->

# Application Answer Bank

Every screening question the agent is allowed to answer without asking. Anything not in
here → **stop and ask the user**, then append the answer to this file so it is never asked
twice.

`TODO` means unanswered. The agent must not invent a value for a `TODO`.

---

## Identity & contact
| Field | Value |
|---|---|
| Full name | TODO |
| Email | TODO |
| Phone | TODO |
| Street address | TODO |
| City / State / ZIP | TODO |
| Country | TODO |
| LinkedIn | see `jobs/input/profile/links.md` |
| GitHub | see `jobs/input/profile/links.md` |
| Portfolio / website | see `jobs/input/profile/links.md` |

## Work authorization  ← most common blocker; fill these first
| Question | Answer |
|---|---|
| Are you legally authorized to work in the US? | TODO |
| Will you now or in the future require sponsorship? | TODO |
| Visa status (if relevant) | TODO |
| Do you hold a security clearance? | No |

## Logistics
| Question | Answer |
|---|---|
| Willing to relocate? | TODO |
| Preferred work mode (remote/hybrid/onsite) | TODO |
| Willing to travel? (FDE roles often ask 25–50%) | TODO |
| Notice period / earliest start date | TODO |
| Current location & time zone | TODO |

## Compensation
| Question | Answer |
|---|---|
| Desired base salary | TODO |
| Salary expectation phrasing when a single number is forced | TODO |
| Current salary (illegal to ask in many states — leave blank if optional) | Prefer not to say |

## History
| Question | Answer |
|---|---|
| Have you previously worked for this company? | No — *(verify per company)* |
| Do you have any non-compete restrictions? | TODO |
| Were you referred by an employee? | No — *(override per application if yes)* |
| How did you hear about this role? | Company website / job board — *(match the source)* |

## Voluntary self-identification (EEO)
The user's choice, applied consistently. These are always optional on US applications.

| Question | Answer |
|---|---|
| Gender | TODO — or "Decline to self-identify" |
| Race / ethnicity | TODO — or "Decline to self-identify" |
| Veteran status | TODO — or "I don't wish to answer" |
| Disability status | TODO — or "I don't wish to answer" |

> Default if the user has not specified: choose the "decline / prefer not to answer" option.
> Never guess an actual demographic value.

---

## Free-text answers

### "Why do you want to work here?"
Not a stored answer — write it fresh per company in `/job tailor`, grounded in that
company's actual product. 3–5 sentences. Never generic.

### "Tell us about a project you're proud of"
Default source: the strongest relevant project in `jobs/input/profile/master-resume.md`. Reshape for
the JD; do not invent outcomes.

### Standard short answers
| Prompt | Answer |
|---|---|
| Years of Python experience | TODO |
| Years of professional software experience | TODO |
| Years working with LLMs / GenAI | TODO |
| Highest degree earned | TODO |
| Are you comfortable with customer-facing work? | TODO |

---

## Changelog
Append `YYYY-MM-DD — added <question> (asked during <job-id>)` when a new answer is learned.
