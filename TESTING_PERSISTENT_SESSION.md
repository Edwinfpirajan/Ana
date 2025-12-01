# 🧪 Testing Persistent Session Mode

## How to Test Persistent Session

### Setup
1. Make sure Ollama and Whisper are running
2. Start Ana: `./ana.exe`
3. Wait for "Ana Streamer Active" message

### Test Flow

**Step 1: Activate Session**
- Say: "Ana"
- Expected: Ana responds and enters listening state
- Log should show: `Pipeline state: listening`

**Step 2: First Command**
- Say any command (e.g., "what time is it")
- Expected: Ana processes and responds
- Log should show: `Pipeline state: recording` → `processing` → `listening`
- ⚠️ **IMPORTANT**: Ana should return to `StateListening`, not `StateIdle`

**Step 3: Second Command (without repeating "Ana")**
- Say another command (e.g., "tell me a joke")
- Expected: Ana processes second command without needing "Ana" prefix
- Log should show same state transitions

**Step 4: Continue Session**
- Repeat Step 3 multiple times
- Ana should continue listening and processing commands

**Step 5: Deactivation**
- Say deactivation word: "Adiós" or "Silencio" or "Detente"
- Expected: Ana responds with goodbye message and returns to idle
- Log should show: `Pipeline state: idle`
- Ana should stop listening for new commands

### Expected Log Sequence

For each command after activation:
```
Pipeline state: listening     ← Waiting for next command
Pipeline state: recording     ← Audio detected
Pipeline state: processing    ← Processing speech
Transcribed text=...          ← STT result
Response=...                  ← LLM response
Pipeline state: listening     ← IMPORTANT: Back to listening, not idle
```

### Deactivation Log Sequence

When deactivation word is spoken:
```
Pipeline state: processing
Deactivation word detected
Pipeline state: idle          ← Session ends
```

### What NOT to Expect

- ❌ Should NOT return to `StateIdle` after each command
- ❌ Should NOT require saying "Ana" for every command
- ❌ Should NOT exit session after first command

## Troubleshooting

**Problem: Ana stops listening after first command**
- Check logs for `Pipeline state: idle`
- This means deactivation check failed or state management is broken
- Solution: Check `processRecordedAudio()` function

**Problem: Commands not being processed**
- Check if Ana is in `StateListening` after first command
- If stuck in `StateRecording`, there's an audio capture issue
- Solution: Verify PortAudio and microphone are working

**Problem: Deactivation doesn't work**
- Say deactivation word clearly: "Adiós Ana"
- Check logs for `Deactivation word detected`
- If not detected, word may need to be added to `IsAnaDeactivated()` function

## Success Criteria

✅ Session activates with "Ana"
✅ Multiple commands processed without "Ana" prefix
✅ Session deactivates with deactivation word
✅ No crashes or errors
✅ State transitions: listening → recording → processing → listening
