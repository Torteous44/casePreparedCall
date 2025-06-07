# SessionState Object Schema

Each `SessionState` object tracks the full runtime context of an interview session. It represents the evolving state of a conversation between the candidate and the AI case interviewer. It is continously running and analyzes the transcript where the current interview is.

## Required Fields

### `current_question`
- **Type:** `integer`
- **Default:** `0`
- **Purpose:** Tracks which question in the `Lesson.questions[]` array is currently active.

---

### `silence_timer`
- **Type:** `integer` (seconds)
- **Default:** `0`
- **Purpose:** Tracks how long the user has been silent.
- **Behavior:**  
  - **IF** `silence_timer >= 7s`  
    → Ask: “Do you need more time or would you like a hint?”  
  - **IF** user responds with **"more time"** or something similar.
    → Wait ~12s more  
    → Ask again if they want a hint or more time  
  - **IF** user responds **"hint"**  
    → Provide the next unused hint from the `Question.hints[]` array  
  - **IF** user remains silent after second prompt  
    → Provide a hint automatically

---

### `hints_used`
- **Type:** `integer`
- **Default:** `0`
- **Purpose:** Counts the number of hints used on the current question.
- **Behavior:**  
  - Increments when a hint is given (voluntarily or due to silence)
  - Cap: maximum depends on the number of available `Question.hints[]`.

---

### `components_hit`
- **Type:** `array of strings`
- **Default:** `[]`
- **Purpose:** Stores `expected_components.title` identifiers the user has mentioned in their response.
- **Used For:**  
  - Evaluating answer quality tier:  
    - Poor: `<2 components hit`  
    - Satisfactory: `2–3 components hit`  
    - High: `≥4 components hit`  
  - Drives prompting logic based on completeness.

---

### `steps_hit`
- **Type:** `array of strings`
- **Default:** `[]`
- **Purpose:** Tracks `GuideStep.label` identifiers that the user has structurally completed.
- **Used For:**  
  - Determining which `GuideStep.clarifier_prompt` to issue
  - Ensuring format quality of response, separate from content

---

### `follow_ups_used`
- **Type:** `array of integers`
- **Default:** `[]`
- **Purpose:** Stores indices of follow-up prompts from `Question.follow_ups[]` that have been triggered
- **Cap:** typically 1–2 per question

---

### `user_ready`
- **Type:** `boolean`
- **Default:** `false`
- **Purpose:** Tracks whether user has confirmed readiness to proceed to next question.

---

### `user_ready_question`
- **Type:** `string`
- **Default:** `"Are you ready to move on to the next question?"`
- **Behavior:** Asked at the end of every question (after response is rated).

---

### `completed`
- **Type:** `boolean`
- **Default:** `false`
- **Purpose:** Indicates whether the current question has been completed (not the entire case).
- **Logic for completion:**
  - **High-quality answer:**  
    → Ask `user_ready_question`  
    → If yes → set `completed = true`
  - **Satisfactory answer:**  
    → Ask `user_ready_question`  
    → If no → ask: “Is there anything else you’d like to add?” or “Would you like a little guidance?”  
    → If follow-up satisfied → ask `user_ready_question` again  
    → If yes → set `completed = true`
  - **Poor-quality answer:**  
    → Trigger a targeted `GuideStep.clarifier_prompt`  
    → If still poor → deliver a hint  
    → Re-score  
    → Continue loop until: either improved, hint cap hit, or user agrees to move on

---

# Behavioral Logic Overview

## Silence Detection

