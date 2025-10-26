"""
Main SICA evolution loop - Algorithm 1 from paper
Self-Referential Agent Improvement
"""

import subprocess
import sys
import importlib
import requests
from archive import Archive
from benchmarks import Benchmark
from config import LLM_BASE_URL, LLM_MODEL, LLM_API_KEY, N_ITERATIONS


class EvolutionLoop:
    def __init__(self, model=None):
        self.model = model or LLM_MODEL
        self.base_url = LLM_BASE_URL
        self.api_key = LLM_API_KEY
        self.archive = Archive()
        self.benchmark = Benchmark()
        self.agent_file = 'agent.py'
    
    def run(self, n_iterations=N_ITERATIONS):
        """
        Main evolution loop - implements Algorithm 1 from SICA paper
        
        Algorithm:
        1. Initialize agent A_0
        2. For i = 0 to n-1:
           3. Evaluate A_i on benchmarks B → score p_i
           4. Run A_best to generate A_(i+1)
        5. Return A_n
        """
        print("=" * 70)
        print("SELF-EVOLUTION STARTING")
        print("Algorithm 1: Self-Referential Agent Improvement (SICA)")
        print("=" * 70)
        
        for i in range(n_iterations):
            print(f"\n{'='*70}")
            print(f"ITERATION {i}")
            print(f"{'='*70}")
            
            # Step 1: Load current agent code
            with open(self.agent_file, 'r') as f:
                current_code = f.read()
            
            # Step 2: Evaluate current agent
            score = self._evaluate_agent()
            print(f"\n📊 Current Score: {score:.2%}")
            
            # Step 3: Store in archive
            self.archive.add_iteration(i, current_code, score)
            
            # Step 4: Meta-improvement (if not last iteration)
            if i < n_iterations - 1:
                print("\n" + "="*70)
                print("META-IMPROVEMENT PHASE")
                print("="*70)
                self._meta_improve()
            
            print(f"\n✓ Iteration {i} complete. Score: {score:.2%}")
        
        # Final summary
        best = self.archive.get_best()
        print("\n" + "=" * 70)
        print("EVOLUTION COMPLETE")
        print("=" * 70)
        print(f"Best Score: {best['score']:.2%} (Iteration {best['iteration']})")
        print(f"Total Iterations: {len(self.archive.iterations)}")
        print("=" * 70)
    
    def _evaluate_agent(self):
        """Evaluate current agent on benchmarks"""
        # Reload agent module to get latest version
        if 'agent' in sys.modules:
            importlib.reload(sys.modules['agent'])
        
        from agent import SelfEvolvingAgent
        agent = SelfEvolvingAgent(self.model)
        
        # Run benchmarks
        score = self.benchmark.evaluate(agent)
        return score
    
    def _meta_improve(self):
        """
        The key SICA step: best agent improves itself
        """
        # 1. Get archive summary
        archive_summary = self.archive.get_summary()
        
        # 2. Get current code
        with open(self.agent_file, 'r') as f:
            current_code = f.read()
        
        # 3. Ask Ollama to propose improvement
        prompt = f"""You are improving a self-evolving agent.

ARCHIVE SUMMARY:
{archive_summary}

CURRENT AGENT CODE:
```python
{current_code}
```

TASK:
Analyze the code and propose ONE specific improvement that would likely increase benchmark scores.

The agent needs to:
1. Better understand tasks in _think()
2. Generate valid JSON plans in _plan()
3. Execute actions correctly in _execute_plan()

Focus on:
- Better reasoning logic
- More robust JSON parsing
- Better error handling
- Improved tool usage

Provide the COMPLETE modified agent.py file.
Return ONLY the Python code, no explanations, no markdown blocks."""
        
        print("\n🤖 Asking Ollama for improvement...")
        improved_code = self._call_ollama(prompt)
        
        # 4. Extract code from response
        improved_code = self._extract_code(improved_code)
        
        # 5. Validate syntax
        if self._validate_code(improved_code):
            print("✓ Generated code has valid syntax")
            
            # 6. Git commit current version (backup)
            subprocess.run(['git', 'add', self.agent_file], capture_output=True)
            subprocess.run(['git', 'commit', '-m', 'Pre-improvement checkpoint'], capture_output=True)
            
            # 7. Apply improvement
            with open(self.agent_file, 'w') as f:
                f.write(improved_code)
            
            print("✓ Improvement applied to agent.py")
            
            # 8. Quick test
            print("\n🧪 Testing improvement...")
            new_score = self._evaluate_agent()
            
            best_so_far = self.archive.get_best()['score']
            
            if new_score >= best_so_far:
                print(f"✓ IMPROVEMENT SUCCESSFUL! {best_so_far:.2%} → {new_score:.2%}")
                subprocess.run(['git', 'add', self.agent_file], capture_output=True)
                subprocess.run(['git', 'commit', '-m', f'Improvement accepted: {new_score:.2%}'], capture_output=True)
            else:
                print(f"✗ No improvement: {new_score:.2%} < {best_so_far:.2%}")
                print("⏮  Reverting to previous version...")
                subprocess.run(['git', 'reset', '--hard', 'HEAD~1'], capture_output=True)
        else:
            print("✗ Generated code has syntax errors. Skipping improvement.")
    
    def _call_ollama(self, prompt):
        """Call local LLM via OpenAI-compatible API"""
        try:
            response = requests.post(
                f"{self.base_url}/chat/completions",
                headers={
                    "Content-Type": "application/json",
                    "Authorization": f"Bearer {self.api_key}"
                },
                json={
                    "model": self.model,
                    "messages": [
                        {"role": "user", "content": prompt}
                    ],
                    "temperature": 0.7,
                    "max_tokens": 4096
                },
                timeout=300  # 5 minute timeout
            )
            response.raise_for_status()
            result = response.json()
            return result['choices'][0]['message']['content']
        except requests.exceptions.Timeout:
            print("⚠ LLM API call timed out")
            return ""
        except Exception as e:
            print(f"⚠ LLM API error: {e}")
            return ""
    
    def _extract_code(self, response):
        """Extract Python code from markdown or raw response"""
        # Try to find code blocks
        if '```python' in response:
            start = response.find('```python') + 9
            end = response.find('```', start)
            return response[start:end].strip()
        elif '```' in response:
            start = response.find('```') + 3
            end = response.find('```', start)
            return response[start:end].strip()
        else:
            # Assume it's all code
            return response.strip()
    
    def _validate_code(self, code):
        """Check if code has valid Python syntax"""
        try:
            compile(code, '<string>', 'exec')
            return True
        except SyntaxError as e:
            print(f"Syntax error: {e}")
            return False


if __name__ == "__main__":
    print("\n🚀 Starting Self-Evolution Loop")
    print(f"📝 Using model: {LLM_MODEL}")
    print(f"🌐 API endpoint: {LLM_BASE_URL}")
    print(f"🔄 Iterations: {N_ITERATIONS}\n")
    
    loop = EvolutionLoop()
    loop.run(n_iterations=N_ITERATIONS)