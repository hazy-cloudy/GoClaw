---
name: pptx
description: Use this skill when the user asks to create, edit, review, or repair PowerPoint `.pptx` presentations. It supports both generating decks from scratch and making surgical changes to existing files.
---

# PPTX

Use this skill when the task is specifically about PowerPoint `.pptx` files.

This workspace install is adapted from Anthropic's `pptx` skill so it works locally inside GoClaw without requiring a global install.

## When To Use It

- Creating a new presentation from scratch
- Editing slide text, notes, order, or structure
- Reviewing an existing deck and proposing changes
- Repairing or inspecting a broken `.pptx`
- Generating helper artifacts such as thumbnails or unpacked XML

## Core Principles

- Prefer `pptxgenjs` when building a deck from scratch or doing broad layout changes.
- Prefer unpacking the `.pptx` archive when you need precise edits to an existing presentation.
- Always keep generated files inside the workspace.
- After structural edits, create a preview image when possible so the result can be checked quickly.

## Workflow

### 1. Understand The Goal

Clarify whether the user wants one of these:

- Create a new deck
- Modify an existing deck
- Review and comment on a deck
- Repair a broken deck

For new deck requests, use this interaction gate before generating slides.

Do not generate a full deck yet if any of these are still unclear:

- Topic or main content
- Audience or use case
- Page count or expected depth
- Whether source material already exists

When one or more of the items above are missing, stop and ask concise clarification questions first.

If useful, also ask about style, language, speaking time, or whether notes are needed, but only when those details clearly affect the output.

Rules:

- Ask at most 3 to 4 concise questions in one turn.
- Prefer one compact grouped message instead of many back-and-forth turns.
- Do not repeat questions if the user already provided the answer.
- If the request is vague like "make a PPT" or "help me do slides", default to clarification first, not generation first.
- If topic, audience, page count, or source material are missing, do not generate a complete slide-by-slide deck yet.
- If the user explicitly wants a quick draft, ask fewer questions and proceed with a rough first version.
- For edits to an existing deck, ask only about the change scope and whether the current layout should be preserved.
- For edits to an existing deck, do not ask the full new-deck question set unless the user is effectively asking for a rebuild.
- Before generating after clarification, briefly summarize your understanding in one short sentence.
- After finishing a deck, briefly offer follow-up help such as speaker notes, a presentation script, or a slide-by-slide explanation.

Suggested clarification targets:

- What is the deck about?
- Who is it for, or what is the use case?
- How many pages, or should it be brief vs detailed?
- Do you already have materials, or should the deck be drafted from scratch?

### 2. Choose An Editing Path

#### Path A: Build Or Rebuild With `pptxgenjs`

Use this when:

- The user wants a new presentation
- The deck needs large-scale layout changes
- You do not need pixel-perfect preservation of the original file

Reference: `pptxgenjs.md`

#### Path B: Edit The Existing PPTX Package

Use this when:

- The user wants targeted edits to an existing `.pptx`
- The current layout should remain mostly intact
- You need to preserve theme, animations, notes, or embedded assets as much as possible

Reference: `editing.md`

### 3. Validate

After edits:

1. Confirm the `.pptx` opens.
2. Generate a thumbnail preview if the environment supports it:

```bash
python workspace/skills/pptx/scripts/thumbnail.py path/to/deck.pptx
```

3. If you unpacked the file for editing, remove temp folders when done:

```bash
python workspace/skills/pptx/scripts/clean.py path/to/unpacked-dir
```

## Helper Commands

Unpack a presentation:

```bash
python workspace/skills/pptx/scripts/office/unpack.py path/to/deck.pptx
```

Pack an unpacked folder back into a `.pptx`:

```bash
python workspace/skills/pptx/scripts/office/pack.py path/to/unpacked-dir path/to/output.pptx
```

Clone an existing slide as a starting point for a new slide:

```bash
python workspace/skills/pptx/scripts/add_slide.py path/to/deck.pptx --source-index 1 --output path/to/output.pptx
```

## Constraints

- Do not overwrite the user's source file unless they explicitly ask for in-place edits.
- Keep intermediate files in the workspace.
- Prefer reversible changes and save edited output as a new file when possible.
- If preview generation is unavailable on the current machine, explain that clearly instead of pretending validation succeeded.
