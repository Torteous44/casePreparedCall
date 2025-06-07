# Example Session Object for Premier Oil Case

This is a complete example of the session initialization object that should be passed from the frontend to the backend to start an interview session.

## API Endpoint
`POST /api/interview/init-with-lesson`

## Example Request Body

```json
{
  "lesson": {
    "lesson_id": "premier_oil_profitability_case_2021",
    "case_id": "Premier Oil",
    "case_image": "",
    "case_type": "Profitability",
    "case_level": "Easy",
    "case_company": "McKinsey & Company",
    "case_description": "2021 McKinsey inspired case about profitability challenges facing a UK offshore upstream oil producer.",
    "case_prompt": "The pandemic-induced collapse in oil prices sharply reduced profitability of Premier Oil, a major UK-based offshore upstream oil and gas producer operating rigs in seven North Sea areas. The CEO has asked your team to design a profitability-improvement plan.",
    "case_prompt_additional_information": "The client has assets only in the North Sea and doesn't plan to adjust its asset portfolio. The profitability for 2020 was -12% (losses), which was common in the industry that year. There is no specific goal to improve profitability. The client is an independent oil and gas company owned by a wide variety of strategic investors.",
    "questions": [
      "What factors would you consider to work on this problem?",
      "Given there is not much Premier Oil can do to increase sales, the manager wants us to focus on costs. To begin with, what are Premier Oil's major expenses?",
      "Maintenance costs have been increasing for Premier Oil's offshore platforms. What might be the reasons behind this?",
      "Retrofitting the existing equipment might reduce some costs. Can you calculate what cost savings the client will be able to capture?"
    ],
    "case_introduction": "premier_oil_profitability_case_2021.Introduction",
    "case_conclusion": "premier_oil_profitability_case_2021.Conclusion"
  },
  "introduction": {
    "introduction_id": "premier_oil_profitability_case_2021.Introduction",
    "introduction_case_prompt": "The pandemic-induced collapse in oil prices sharply reduced profitability of Premier Oil, a major UK-based offshore upstream oil and gas producer. Premier Oil operates rigs in seven areas in the North Sea. The CEO has brought your team in to design a profitability improvement plan.",
    "introduction_additional_information": "The client has assets only in the North Sea and doesn't plan to adjust its asset portfolio. The profitability for 2020 was -12% (losses), which was common in the industry that year. There is no specific goal to improve profitability. The client is an independent oil and gas company owned by a wide variety of strategic investors.",
    "introduction_guide_steps": "premier_oil_profitability_case_2021.IntroductionGuideSteps",
    "introduction_question_prompt": "Do you have any questions about the case?"
  },
  "questions": [
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
        "Have you considered analyzing Premier Oil's specific clients and products to understand what drives their performance?",
        "Could it help to break the company's financials down into revenue and fixed versus variable cost structure?",
        "Have you thought about which levers Premier Oil could pull to either boost revenues or cut costs in the short and long term?"
      ],
      "follow_ups": [
        "How would you benchmark Premier Oil's performance against typical players in the upstream oil and gas sector?",
        "What do you know about Premier Oil's clients, and how might its product mix affect profitability?",
        "Can you go deeper into how fixed and variable costs play out in their cost structure?",
        "Let's talk about cost-saving ideas. Where exactly might you streamline or renegotiate to improve profitability?"
      ],
      "clarifiers": [
        "Are you referring to upstream or downstream oil companies? Can you clarify the type of comparison?",
        "When you mention customers, do you mean institutional buyers or end consumers? Please clarify.",
        "Your breakdown of costs was a bit unclear—can you specify what falls under fixed vs. variable?",
        "You mentioned 'improving processes'—can you be more specific about what type of revenue or cost lever that is?"
      ]
    },
    {
      "question_id": "premier_oil_profitability_case_2021.Q2",
      "question_prompt": "Given there is not much Premier Oil can do to increase sales, the manager wants us to focus on costs. To begin with, what are Premier Oil's major expenses?",
      "expected_components": [
        {
          "title": "Operational costs",
          "description": "Day-to-day operational expenses including labor, utilities, and consumables."
        },
        {
          "title": "Maintenance costs",
          "description": "Regular and preventive maintenance of offshore platforms and equipment."
        },
        {
          "title": "Capital expenditures",
          "description": "Investment in new equipment, platform upgrades, and infrastructure."
        },
        {
          "title": "Regulatory and compliance costs",
          "description": "Environmental compliance, safety regulations, and industry standards."
        }
      ],
      "guide_steps": "premier_oil_profitability_case_2021.Q2GuideSteps",
      "hints": [
        "Think about the major cost categories for an offshore oil production company.",
        "Consider both fixed and variable costs in your analysis.",
        "What about regulatory and environmental compliance costs?",
        "Don't forget about maintenance costs for offshore platforms."
      ],
      "follow_ups": [
        "Which of these cost categories do you think offers the most potential for savings?",
        "How might the offshore nature of their operations affect their cost structure?",
        "What role does equipment age play in maintenance costs?",
        "How do regulatory costs compare to operational costs in this industry?"
      ],
      "clarifiers": [
        "Can you be more specific about what operational costs include?",
        "Are you referring to routine or emergency maintenance?",
        "What exactly do you mean by capital expenditures in this context?",
        "Can you clarify the difference between regulatory and operational costs?"
      ]
    },
    {
      "question_id": "premier_oil_profitability_case_2021.Q3",
      "question_prompt": "Maintenance costs have been increasing for Premier Oil's offshore platforms. What might be the reasons behind this?",
      "expected_components": [
        {
          "title": "Equipment aging",
          "description": "Older equipment requires more frequent and expensive maintenance."
        },
        {
          "title": "Harsh environmental conditions",
          "description": "Offshore platforms face challenging weather and corrosive saltwater environment."
        },
        {
          "title": "Deferred maintenance",
          "description": "Previous cost-cutting may have resulted in delayed maintenance, leading to higher costs now."
        },
        {
          "title": "Supply chain issues",
          "description": "Remote offshore location makes parts and specialist services more expensive."
        }
      ],
      "guide_steps": "premier_oil_profitability_case_2021.Q3GuideSteps",
      "hints": [
        "Consider the age of Premier Oil's offshore infrastructure.",
        "Think about the unique challenges of offshore operations.",
        "What impact might previous cost-cutting have had?",
        "How does location affect maintenance costs and logistics?"
      ],
      "follow_ups": [
        "How would you prioritize addressing these maintenance cost drivers?",
        "What role does predictive maintenance play in offshore operations?",
        "How might weather patterns affect maintenance scheduling and costs?",
        "What are the trade-offs between preventive and reactive maintenance?"
      ],
      "clarifiers": [
        "Are you referring to mechanical wear or environmental damage?",
        "Can you clarify what you mean by deferred maintenance?",
        "Are you talking about supply chain costs or availability issues?",
        "What specific environmental factors are you considering?"
      ]
    },
    {
      "question_id": "premier_oil_profitability_case_2021.Q4",
      "question_prompt": "Retrofitting the existing equipment might reduce some costs. Can you calculate what cost savings the client will be able to capture?",
      "expected_components": [
        {
          "title": "Data requirements",
          "description": "Need specific cost data: current maintenance costs, retrofitting costs, expected savings."
        },
        {
          "title": "Calculation framework",
          "description": "Net present value analysis considering initial investment vs. ongoing savings."
        },
        {
          "title": "Risk factors",
          "description": "Operational risks, implementation timeline, and uncertainty in savings estimates."
        },
        {
          "title": "Implementation considerations",
          "description": "Timing, resource requirements, and potential operational disruptions."
        }
      ],
      "guide_steps": "premier_oil_profitability_case_2021.Q4GuideSteps",
      "hints": [
        "What specific data would you need to perform this calculation?",
        "Think about the time value of money in your analysis.",
        "Consider both the costs and benefits of retrofitting.",
        "What assumptions would you need to make?"
      ],
      "follow_ups": [
        "How would you approach gathering the necessary data?",
        "What discount rate would be appropriate for this analysis?",
        "How would you account for operational risks in your calculation?",
        "What sensitivity analysis would you perform on your estimates?"
      ],
      "clarifiers": [
        "Are you looking for a specific formula or a general approach?",
        "What time horizon are you considering for the analysis?",
        "Are you including implementation costs in your calculation?",
        "How are you defining 'cost savings' in this context?"
      ]
    }
  ],
  "guide_steps": {
    "premier_oil_profitability_case_2021.Q1GuideSteps": {
      "guide_steps": [
        {
          "step_id": 1,
          "label": "Step 1: Do horizontal presentation",
          "description": "The best practice is to start with a 15-second big-picture overview, e.g., "I'd like to assess this problem through the lens of four areas – first, …; secondly, …; thirdly, …; and finally …"",
          "clarifier_prompt": ""
        },
        {
          "step_id": 2,
          "label": "Step 2: Hit key points",
          "description": "Check if the candidate covers all key points typical for a profitability case structure:\n- Profitability analysis\n  • Revenue analysis\n  • Cost structure\n- Business model\n- External factors\n  • Client segments\n  • Growth rate*\n  • Product portfolio\n  • Competition*\n  • Typical margin\n\n* = less important in this case, as the prompt hints that the pandemic was the root cause.",
          "clarifier_prompt": "Can you walk me through the key components you'd consider when evaluating a company's profitability?"
        },
        {
          "step_id": 3,
          "label": "Step 3: Add stories (optional)",
          "description": "To avoid a cookie-cutter or generic approach, the candidate can incorporate 2–3 stories or industry-specific insights into their structure, e.g.:\n- "It's a capex-heavy business, so economies of scale are crucial."\n- "Crude oil is a commodity highly dependent on global markets, so we don't determine our pricing strategy much."\n- "Offshore platforms are likely subject to strict environmental regulation which might manifest in higher costs."",
          "clarifier_prompt": "Are there any industry-specific insights or contextual stories you might add to strengthen your analysis?"
        },
        {
          "step_id": 4,
          "label": "Step 4: Finish with a question",
          "description": "At the end of the structure presentation, it is helpful for the candidate to show initiative and forward momentum, e.g., "If this approach sounds reasonable, I'd like to start by digging into financials. Do we have revenue data?"",
          "clarifier_prompt": "What question would you ask to move the case forward if your structure seems sound?"
        }
      ]
    }
  },
  "conclusion": {
    "conclusion_id": "premier_oil_profitability_case_2021.Conclusion",
    "farewell_script": "Thank you for your time today, [user_name].",
    "next_steps_script": "We'll be keeping in touch regarding next steps.",
    "post_case_question_response": "The interview has concluded. Thank you so much for your time again."
  },
  "persona": {
    "case_interview_company": "McKinsey & Company",
    "interviewer_tone": "neutral and analytical",
    "greeting_style": "scripted intro with case context and light warm-up",
    "general_persona": "You are a case interviewer representing McKinsey & Company. You maintain a logical, efficient, and professional tone throughout the interview. You guide the candidate through a structured business case, asking follow-ups as needed. You provide minimal praise and focus on evaluating clarity of thought, structure, and communication. You begin with a scripted greeting and focus on simulating the feel of a real McKinsey case interview."
  },
  "sample_rate": 16000,
  "encoding": "pcm_s16le"
}
```

## Expected Response

```json
{
  "session_id": "uuid-generated-session-id",
  "websocket_url": "ws://localhost:8080/ws/interview/uuid-generated-session-id",
  "status": "initialized"
}
```

## Frontend Integration

The frontend should send this complete object to initialize a session, then connect to the returned WebSocket URL to begin audio streaming. The backend will have access to all lesson components for context-aware interview management.

## Context Brain Access

Once initialized, the Context Brain will have access to:

- **Static Context**: Complete lesson definition with all linked objects
- **Dynamic Context**: Session state and real-time transcript
- **Behavioral Logic**: Silence detection, hint management, progress tracking
- **Response Generation**: Persona-appropriate interviewer responses 