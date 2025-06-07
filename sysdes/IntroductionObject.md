# Introduction Object Schema

This section defines the fields for an `Introduction` object used within a `Lesson`.

## Fields

- `introduction_id`: *string* — Unique identifier composed as `"<lesson_id>.Introduction"`.
- `introduction_case_prompt`: *string* — Full background case prompt, delivered before the case questions begin.
- `introduction_additional_information`: *string* — Any supporting background info (optional; can be left blank).
- `introduction_guide_steps`: *string* — Reference to the corresponding guide_steps object for expected answer structure.
- `introduction_question_prompt`: *string* — Question the AI asks after delivering the case prompt (usually to confirm understanding or elicit clarifying questions).

---

# Example Introduction Object: Premier Oil

```json
{
  "introduction_id": "premier_oil_profitability_case_2021.Introduction",
  "introduction_case_prompt": "The pandemic-induced collapse in oil prices sharply reduced profitability of Premier Oil, a major UK-based offshore upstream oil and gas producer. Premier Oil operates rigs in seven areas in the North Sea. The CEO has brought your team in to design a profitability improvement plan.",
  "introduction_additional_information": "The client has assets only in the North Sea and doesn’t plan to adjust its asset portfolio. The profitability for 2020 was -12% (losses), which was common in the industry that year. There is no specific goal to improve profitability. The client is an independent oil and gas company owned by a wide variety of strategic investors.",
  "introduction_guide_steps": "premier_oil_profitability_case_2021.IntroductionGuideSteps",
  "introduction_question_prompt": "Do you have any questions about the case?"
}
