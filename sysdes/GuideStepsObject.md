# GuideSteps Object Schema

Each `GuideSteps` object defines the expected structural format a candidate should follow when answering a given case interview question. These are meant to simulate the internal logic of a strong, MBB-style structured response. There is a `GuideSteps` object associated with every `QuestionObject` and every `IntroductionObject`

## Fields

Each guide step is an object with the following fields:

- `step_id`: *integer* — A number indicating this step’s position in the guiding structure (typically from 1 to 4).
- `label`: *string* — A title describing what this step accomplishes (e.g., “Step 2: Hit key points”). May include an `(optional)` flag.
- `description`: *string* — A detailed explanation of what this step entails and how the candidate should approach it.
- `clarifier_prompt`: *string (optional)* — A prompt the AI interviewer uses if the candidate fails to complete this step.

---

# Example GuideSteps Object: Premier Oil – Question #1

```json
{
  "guide_steps": [
    {
      "step_id": 1,
      "label": "Step 1: Do horizontal presentation",
      "description": "The best practice is to start with a 15-second big-picture overview, e.g., “I’d like to assess this problem through the lens of four areas – first, …; secondly, …; thirdly, …; and finally …”",
      "clarifier_prompt": ""
    },
    {
      "step_id": 2,
      "label": "Step 2: Hit key points",
      "description": "Check if the candidate covers all key points typical for a profitability case structure:\n- Profitability analysis\n  • Revenue analysis\n  • Cost structure\n- Business model\n- External factors\n  • Client segments\n  • Growth rate*\n  • Product portfolio\n  • Competition*\n  • Typical margin\n\n* = less important in this case, as the prompt hints that the pandemic was the root cause.",
      "clarifier_prompt": "Can you walk me through the key components you’d consider when evaluating a company’s profitability?"
    },
    {
      "step_id": 3,
      "label": "Step 3: Add stories (optional)",
      "description": "To avoid a cookie-cutter or generic approach, the candidate can incorporate 2–3 stories or industry-specific insights into their structure, e.g.:\n- “It’s a capex-heavy business, so economies of scale are crucial.”\n- “Crude oil is a commodity highly dependent on global markets, so we don’t determine our pricing strategy much.”\n- “Offshore platforms are likely subject to strict environmental regulation which might manifest in higher costs.”",
      "clarifier_prompt": "Are there any industry-specific insights or contextual stories you might add to strengthen your analysis?"
    },
    {
      "step_id": 4,
      "label": "Step 4: Finish with a question",
      "description": "At the end of the structure presentation, it is helpful for the candidate to show initiative and forward momentum, e.g., “If this approach sounds reasonable, I’d like to start by digging into financials. Do we have revenue data?”",
      "clarifier_prompt": "What question would you ask to move the case forward if your structure seems sound?"
    }
  ]
}
