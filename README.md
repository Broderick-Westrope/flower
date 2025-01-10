# Flower

Flower (pronounced [flow]-[ah]) is a CLI tool to assist you in using the flowtime time-management technique, as described by Zoë Read-Bivens [here](https://medium.com/@UrgentPigeon/the-flowtime-technique-7685101bd191).

## Install

```sh
go install github.com/Broderick-Westrope/flower
```

## Usage

```sh
# Interactively add a new task and start a session
flower start
# Enter your desired task details

# Quickly add a new task
flower task add --name "Learn touch typing" --desc "Cause I gotta be fast"
# Output:
# New task added with ID 2.

# In case you forget the task IDs you can list all tasks
flower task list
# Output:
# 1. Name: Solve Rubik's cube
# 2. Name: Learn touch typing
#    Description: Cause I gotta be fast

# Start a new session for the task with ID 2, skipping the interactive check for open sessions
flower start 2 --skip
# Output:
# Session started for task "Learn touch typing".

# List all open sessions
flower list --open
# Output:
# 2 sessions found:
#   - Task: "Solve Rubik's cube",
#     Duration: 2m5.31859s,
#     State: Open
#   - Task: "Learn touch typing",
#     Duration: 22.3186s,
#     State: Open

# Stop the latest session and get suggested break
flower stop --latest
# Output:
# Stopping the latest open session.
# Latest session stopped. The suggested break is 5 minutes.

# More advanced data anaylysis can be done with jq (https://github.com/jqlang/jq).
# The following are examples that require jq to be installed.

# Get the name of the second task
flower list --json | jq '.[1].task.name'
# Output:
# "Learn touch typing"

# Get the name of each closed session (ie. non-null end time)
flower list --json | jq '.[] | select(.ended_at != null) | .task.name'
# Output:
# "Solve Rubik's cube"

# Get tasks which have a description
flower list --json | jq '.[] | select(.task.description != "") | {name: .task.name, description: .task.description}'
# Output:
# {
#   "name": "Learn touch typing",
#   "description": "Cause I gotta be fast"
# }

# Create a summary
flower list --json | jq '{
  total_tasks: length,
  completed: map(select(.ended_at != null)) | length,
  in_progress: map(select(.ended_at == null)) | length,
  tasks: map(.task.name)
}'
# Output:
# {
#   "total_tasks": 2,
#   "completed": 1,
#   "in_progress": 1,
#   "tasks": [
#     "Solve Rubik's cube",
#     "Learn touch typing"
#   ]
# }

# You can do a lot more than this with some time, patience, and saving a few json files. I'm just not a jq wizard ;)
```

## What is flowtime?

For those farmiliar with the pomodoro technique, this is similar. The key difference is that you aren't locked into a time-frame for working. Meaning if you find yourself entering a flow state of mind you can maintain this momentum. In turn for completing larger blocks of productive work you receive a larger break.

If you'd like to learn more I suggest reading the previously linked blog post that Zoë wrote explaining a pen-and-paper approach.

## Why not just use paper?

If you find the paper method (or any other) enough for you then I see no need to continue reading. But, if you're like me and want to better track your time then here are some things I think make `flower` a great tool:

- **Tasks are persistent.** They are stored on your device separately from sessions. Although each session has only one task, a task can have many sessions. If you're working on something big, track your time spent on it using the same task.
- **Data is your friend.** Because a single task is easily linked to multiple session, you can find out:
    - What's your average session length for practicing piano?
    - How often do you skip a break after walking the dog?
    - In the last month how many times have you practiced coding?
    - How often is a work meeting immediately followed by an extended break?
    - What's your most productive time of day?
- **Tasks can be nested.** If you *need* more data you can begin by organising your tasks better. This way you can not only see how much time is spent in meetings, but how much time is spent at work or at your screen, for example.

You've likely noticed that these selling points are all focussed around the data. Although I don't plan to implement all of these data questions directly, the database is unencrypted and stored locally maning you can easily write your own tools to extend the abilities of `flower`, Leveraging computer tools like this is a great way to make inferences into the data that would otherwise be put in the recycling. There are some other, less data-focussed benefits but they are minor:

- **No more math!** I like math, but many don't. You no longer have to calculate the duration of your session or the corresponding break.
- **No more pens.** If you're already on your computer most of your work day (as I am) then this just means a little bit of optimisation.

As I said earlier, if these benefits don't sound appealing then you may be better off using the good old pen-and-paper method.

## Glossary

- **Task:** Anything you'd like to track. It could be writing an essay, playing with your kids, practicing a new skill, or time in the bathroom. If you want to track it, it's a task.
- **Session:** The time spent between starting and stopping the timer. A session is linked to a single task. That is, with flowtime you dedicate your time wholly to a single thing. No multitasking. When you stop focusing on a task, end your session. Switched your focus? Stop this session and start a new one.