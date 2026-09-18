---
name: Question asking style
description: How to phrase every clarifying question — full context, layman language, recommendation with pros/cons
type: preference
---

# Question Asking Style

Apply to EVERY clarifying question without exception. Never ask a bare or context-free question.

## Rules

1. **Full context inline**
   - If the prompt/question is small, include the full prompt inline.
   - If large, include enough surrounding context that a layman understands without external reference.
   - Never assume the user remembers prior turns.

2. **Layman-friendly phrasing**
   - Plain, non-technical language.
   - Define any jargon used in the same question.

3. **Recommendation block (mandatory when choices are involved)**
   - State the recommended option clearly.
   - Explain *why* it is recommended.
   - List **Pros** of the recommendation.
   - List **Cons** of the recommendation.
   - Briefly mention alternatives with their own pros/cons where relevant.

4. **Format**
   - Use `questions--ask_questions` with descriptive `description` fields on each option containing the why/pros/cons summary.
   - The main `question` text should restate the situation in plain English, not just "Which one?".

## How to apply

Before calling `ask_questions`, check:
- [ ] Does the question text stand alone with full context?
- [ ] Is it readable by a non-technical person?
- [ ] Does each option have a recommendation marker, reasoning, pros, and cons?
- [ ] Is my recommended choice clearly identified?

If any box is unchecked, rewrite before sending.
