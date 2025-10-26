# OpenEvolve Configuration

This directory contains configuration files for the OpenEvolve self-evolution system integrated into the Agentic Self-Evolving Command Center.

## What is OpenEvolve?

OpenEvolve is the **reward-based evolution engine** that allows the agent to:
- **Evaluate** its own code execution
- **Reward** successful patterns (+10 to +20 points)
- **Penalize** failures (-10 to 0 points)
- **Evolve** by rewriting its own code based on what works
- **Propagate** intelligence through continuous self-improvement

## Configuration Files

### `agent_config.yaml`
Main configuration for agent evolution:
- Evolution parameters (iterations, population size)
- LLM models (Gemma 3 via Ollama)
- Reward system settings
- Pattern detection thresholds

### `patterns.yaml`
Defines patterns the watchdog monitors:
- Authentication patterns
- API call patterns
- Database query patterns
- Error handling patterns
- Security patterns

### `rewards.yaml`
Reward structure for different outcomes:
- Success rewards (+10 to +20)
- Partial success (+5 to +10)
- Failure penalties (-10 to 0)
- Custom reward functions

### `watchdog.yaml`
Watchdog monitoring configuration:
- Alert severity levels
- Pattern detection sensitivity
- Security checks
- Code quality thresholds

## Integration with Agent

The agent uses OpenEvolve through:

1. **Execution** → Agent runs code/commands
2. **Evaluation** → OpenEvolve scores the result
3. **Reward** → Points assigned based on outcome
4. **Learning** → Patterns stored in Neo4j
5. **Evolution** → Agent rewrites code to improve
6. **Propagation** → Better code becomes the new baseline

## Customization

All YAML files in this directory can be edited to customize:
- What patterns to detect
- How to reward/penalize
- When to trigger alerts
- Which metrics to track

Changes are automatically:
- ✅ Mirrored to Neo4j knowledge graph
- ✅ Visible in the UI (right panel)
- ✅ Used by the agent for learning
- ✅ Tracked in evolution history

## Example Workflow

```yaml
# User asks: "Create a login endpoint"
Agent writes code → 
  OpenEvolve evaluates →
    ✓ Has password hashing: +5
    ✓ Has JWT tokens: +5
    ✓ Has rate limiting: +5
    ✓ Tests pass: +5
    Total: +20 (High reward)
    
Agent stores pattern in Neo4j →
  Next time: "I know how to do auth well!"
```

## Files You Can Edit

| File | Purpose | When to Edit |
|------|---------|--------------|
| `agent_config.yaml` | Core evolution settings | Adjust learning speed, population size |
| `patterns.yaml` | What to detect | Add your domain-specific patterns |
| `rewards.yaml` | How to score | Customize reward structure |
| `watchdog.yaml` | When to alert | Set sensitivity levels |

## Advanced: Custom Patterns

Add your own patterns to `patterns.yaml`:

```yaml
custom_patterns:
  - name: "stripe_integration"
    description: "Detect Stripe payment integration"
    keywords: ["stripe", "payment", "charge"]
    reward: 15
    
  - name: "email_validation"
    description: "Proper email validation"
    regex: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    reward: 5
```

The agent will learn these patterns and apply them automatically!

## See Also

- `/docs/OPENEVOLVE_INTEGRATION.md` - Full integration guide
- `/docs/SELF_MODIFICATION.md` - How the agent rewrites itself
- `/docs/PROPAGATION.md` - Intelligence propagation philosophy

