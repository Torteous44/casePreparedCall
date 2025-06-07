# Lesson Object Schema

This section defines the fields for a single `Lesson` object used in CasePrepared.

## Fields

- `lesson_id`: *string* — Unique identifier for this lesson.
- `case_id`: *string* — Display name of the case (e.g., "Premier Oil").
- `case_image`: *string* — URL or path to image asset (optional).
- `case_type`: *string* — Type of case (e.g., "Profitability", "Market Entry").
- `case_level`: *string* — Difficulty level: one of `"Easy"`, `"Medium"`, `"Difficult"`.
- `case_company`: *string* — One of `"McKinsey & Company"`, `"Bain & Company"`, `"Boston Consulting Group"`.
- `case_description`: *string* — 13–15 word summary of the case background.
- `case_prompt`: *string* — Initial business problem prompt delivered by the AI interviewer.
- `case_prompt_additional_information`: *string* - Additional information that a case has, but only provided on request
- `questions`: *array of strings* — List of 4 interviewer questions objects for the case
- `case_introduction`: *links to a case introduction object* 
- `case_conclusion`: *links to a case conclusion object* 

---

# Example Lesson Object: Premier Oil Profitability

```json
{
  "lesson_id": "premier_oil_profitability_case_2021",
  "case_id": "Premier Oil",
  "case_image": "",  
  "case_type": "Profitability",
  "case_level": "Easy",
  "case_company": "McKinsey & Company",
  "case_description": "2021 McKinsey inspired case about profitability challenges facing a UK offshore upstream oil producer.",
  "case_prompt": "The pandemic-induced collapse in oil prices sharply reduced profitability of Premier Oil, a major UK-based offshore upstream oil and gas producer operating rigs in seven North Sea areas. The CEO has asked your team to design a profitability-improvement plan.",
  "case_prompt_additional_information": "The client has assets only in the North Sea and doesn’t plan to adjust its asset portfolio. The profitability for 2020 was -12% (losses), which was common in the industry that year. There is no specific goal to improve profitability. The client is an independent oil and gas company owned by a wide variety of strategic investors.",
  "questions": [
    "What factors would you consider to work on this problem?",
    "Given there is not much Premier Oil can do to increase sales, the manager wants us to focus on costs. To begin with, what are Premier Oil’s major expenses?",
    "Maintenance costs have been increasing for Premier Oil’s offshore platforms. What might be the reasons behind this?",
    "Retrofitting the existing equipment might reduce some costs. Can you calculate what cost savings the client will be able to capture?"
  ],
  "case_introduction": [INSERT OBJECT],
  "case_conclusion": [INSERT OBJECT]
}
