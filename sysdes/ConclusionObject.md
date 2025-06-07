# Conclusion Object Specification

The **Conclusion** object defines the final wrap-up sequence that occurs after all case questions are complete.  
It personalises the farewell, offers next-steps language, and handles any late-breaking case questions from the candidate.

---

## Fields

| Field | Type | Description |
|-------|------|-------------|
| `conclusion_id` | *string* | Unique identifier, formatted as `"<lesson_id>.Conclusion"`. |
| `farewell_script` | *string* | Main closing line, must include the token `[user_name]` for personalisation. |
| `next_steps_script` | *string* | A brief sentence about keeping in touch or what happens after the interview. |
| `post_case_question_response` | *string* | Fallback line used **only if** the candidate tries to ask additional case-related questions after the interview has ended. |

---

## Behavioural Logic

1. **Primary Closure**  
   - Speak `farewell_script` (token replaced with the candidate’s name).  
   - Follow with `next_steps_script`.

2. **Handling Extra Questions**  
   - **IF** the user asks a *case-related* question after closure  
     → reply with `post_case_question_response`.  
   - **ELSE IF** the user’s question is general career advice or logistics  
     → say I'm sorry, I am an AI interviewer I cannot advise on this.

---

## Example Conclusion Object — Premier Oil

```json
{
  "conclusion_id": "premier_oil_profitability_case_2021.Conclusion",
  "farewell_script": "Thank you for your time today, [user_name].",
  "next_steps_script": "We’ll be keeping in touch regarding next steps.",
  "post_case_question_response": "The interview has concluded. Thank you so much for your time again."
}
