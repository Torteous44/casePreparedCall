# Question Object Schema

Each `Question` object represents a single structured step in a case interview and is linked to its corresponding entry in the `Lesson.questions[]` array.

## Fields

- `question_id`: *string* — Must be a unique identifier in the format `"<lesson_id>.Q<number>"`. Links this object to its corresponding index in `Lesson.questions[]`.
- `question_prompt`: *string* — The actual question the AI interviewer asks the user.
- `expected_components`: *array of objects* — Each expected component includes:
  - `title`: *string* — Label for a key area the candidate should mention.
  - `description`: *string* — Rich context to guide the model's internal evaluation or use in feedback.
- `guide_steps`: *string* — Link to the `GuideSteps` object associated with this question.
- `hints`: *array of strings* — Hints used if the candidate is stuck or misses key areas. Each hint corresponds to one component and uses subtle coaching language.
- `follow_ups`: *array of strings* — Follow-up probes to dig deeper if a component is partially addressed or rushed.
- `clarifiers`: *array of strings* — Clarifying questions used when a candidate gives an ambiguous or incorrect response related to a component.

---

# Example Question Object: Premier Oil – Q1

```json
{
  "question_id": "premier_oil_profitability_case_2021.Q1",
  "question_prompt": "What factors would you consider to work on this problem?",
  "expected_components": [
    {
      "title": "Upstream oil and gas companies",
      "description": "Typical margins, cost structure of several major players for benchmarking, and major trends (apart from pandemic)."
    },
    {
      "title": "Premier Oil",
      "description": "Major accounts (clients), product portfolio (crude oil, gas?), and operational value chain (e.g., extraction, pipeline transportation)."
    },
    {
      "title": "Financial analysis",
      "description": "Revenue analysis and full cost structure, including fixed vs. variable cost distinction."
    },
    {
      "title": "Profitability improvement areas",
      "description": "Opportunities to boost revenue (e.g., secure new contracts) and reduce costs (e.g., optimize fixed or streamline variable costs)."
    }
  ],
  "guide_steps": "premier_oil_profitability_case_2021.Q1GuideSteps",
  "hints": [
    "Have you thought about how profitability benchmarks for upstream oil and gas companies could reveal gaps or opportunities?",
    "Have you considered analyzing Premier Oil’s specific clients and products to understand what drives their performance?",
    "Could it help to break the company’s financials down into revenue and fixed versus variable cost structure?",
    "Have you thought about which levers Premier Oil could pull to either boost revenues or cut costs in the short and long term?"
  ],
  "follow_ups": [
    "How would you benchmark Premier Oil’s performance against typical players in the upstream oil and gas sector?",
    "What do you know about Premier Oil’s clients, and how might its product mix affect profitability?",
    "Can you go deeper into how fixed and variable costs play out in their cost structure?",
    "Let’s talk about cost-saving ideas. Where exactly might you streamline or renegotiate to improve profitability?"
  ],
  "clarifiers": [
    "Are you referring to upstream or downstream oil companies? Can you clarify the type of comparison?",
    "When you mention customers, do you mean institutional buyers or end consumers? Please clarify.",
    "Your breakdown of costs was a bit unclear—can you specify what falls under fixed vs. variable?",
    "You mentioned 'improving processes'—can you be more specific about what type of revenue or cost lever that is?"
  ]
}
