#!/usr/bin/env python3
"""
AI-Powered Changelog Generator for Terraform Provider
Uses OpenAI GPT models for intelligent changelog generation
"""

import os
import sys
import json
import argparse
import re
from datetime import datetime
from typing import List, Dict, Optional
import subprocess

try:
    from openai import OpenAI
    from dotenv import load_dotenv
except ImportError:
    print("Error: Required packages not installed. Run: pip install -r scripts/requirements.txt")
    sys.exit(1)

# Load environment variables
load_dotenv()

class ChangelogGenerator:
    def __init__(self, api_key: str, config_path: str = ".github/changelog-config.json"):
        """Initialize the changelog generator with OpenAI API."""
        self.client = OpenAI(api_key=api_key)
        self.config = self._load_config(config_path)
        
    def _load_config(self, config_path: str) -> Dict:
        """Load configuration from JSON file."""
        try:
            with open(config_path, 'r') as f:
                return json.load(f)
        except FileNotFoundError:
            # Default configuration
            return {
                "ai_model": "gpt-4o",
                "max_commits_per_pr": 100,
                "changelog_style": "terraform-aws-provider",
                "categories": [
                    "BREAKING CHANGES",
                    "FEATURES",
                    "ENHANCEMENTS",
                    "BUG FIXES",
                    "DEPRECATIONS",
                    "DOCUMENTATION"
                ]
            }
    
    def get_commits_from_pr(self, pr_number: Optional[int] = None, 
                           base_ref: str = "HEAD", 
                           head_ref: str = "HEAD^") -> List[Dict]:
        """Fetch commits from git repository."""
        try:
            if pr_number:
                # Get commits from PR
                cmd = f"git log --pretty=format:'%H|||%an|||%ae|||%ad|||%s|||%b' {base_ref}...{head_ref}"
            else:
                # Get commits from commit range
                cmd = f"git log --pretty=format:'%H|||%an|||%ae|||%ad|||%s|||%b' {head_ref}"
            
            result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
            
            if result.returncode != 0:
                print(f"Error fetching commits: {result.stderr}")
                return []
            
            commits = []
            for line in result.stdout.strip().split('\n'):
                if not line:
                    continue
                parts = line.split('|||')
                if len(parts) >= 5:
                    commits.append({
                        'hash': parts[0],
                        'author': parts[1],
                        'email': parts[2],
                        'date': parts[3],
                        'subject': parts[4],
                        'body': parts[5] if len(parts) > 5 else ''
                    })
            
            return commits[:self.config.get('max_commits_per_pr', 50)]
        
        except Exception as e:
            print(f"Error fetching commits: {e}")
            return []
    
    def _preprocess_commits(self, commits: List[Dict]) -> List[Dict]:
        """Filter and score commits by importance."""
        scored_commits = []
        
        for commit in commits:
            score = 0
            subject = commit['subject'].lower()
            body = commit.get('body', '').lower()
            
            # Score based on commit type prefixes (higher priority)
            if subject.startswith('breaking:'):
                score += 10
            elif subject.startswith('deprecate:'):
                score += 9
            elif subject.startswith('feat:'):
                score += 8
            elif subject.startswith('add:'):
                score += 7
            elif subject.startswith('fix:'):
                score += 6
            elif subject.startswith('patch:'):
                score += 6
            elif subject.startswith('enhance:') or subject.startswith('improve:') or subject.startswith('update:'):
                score += 5
            elif subject.startswith('docs:') or subject.startswith('example:'):
                score += 3
            elif subject.startswith('refactor:') or subject.startswith('test:') or subject.startswith('chore:') or subject.startswith('ci:'):
                score += 1  # Low priority, likely to be skipped
            
            # Additional keyword scoring (if no prefix was matched)
            if score == 0:
                if any(kw in subject for kw in ['breaking', 'major', 'removed']):
                    score += 10
                elif any(kw in subject for kw in ['deprecat']):
                    score += 9
                elif any(kw in subject for kw in ['feat', 'feature', 'new']):
                    score += 8
                elif any(kw in subject for kw in ['add']):
                    score += 7
                elif any(kw in subject for kw in ['fix', 'bug', 'patch']):
                    score += 6
                elif any(kw in subject for kw in ['enhance', 'improve', 'update']):
                    score += 5
                elif any(kw in subject for kw in ['doc', 'readme', 'example']):
                    score += 3
                elif any(kw in subject for kw in ['refactor', 'cleanup', 'style', 'test', 'chore']):
                    score += 1
            
            # Boost score for resource/provider changes
            if re.search(r'resource[/_]', subject) or re.search(r'rafay_\w+', subject):
                score += 2
            if re.search(r'data[_\s]source', subject):
                score += 2
            
            commit['importance_score'] = score
            scored_commits.append(commit)
        
        # Sort by importance and return top commits
        scored_commits.sort(key=lambda x: x['importance_score'], reverse=True)
        return scored_commits
    
    def generate_changelog_content(self, commits: List[Dict], 
                                   deprecations: List[Dict],
                                   pr_number: Optional[int] = None,
                                   pr_url: Optional[str] = None) -> str:
        """Generate changelog content using OpenAI GPT."""
        
        # Preprocess commits
        processed_commits = self._preprocess_commits(commits)
        
        # Prepare commit information for AI
        commit_info = []
        for commit in processed_commits:
            commit_info.append(
                f"• {commit['subject']}\n  {commit['body'][:200] if commit.get('body') else ''}\n  (by {commit['author']})"
            )
        
        commits_text = "\n".join(commit_info)
        
        # Prepare deprecation information
        deprecations_text = ""
        if deprecations:
            deprecations_text = "\n\nDEPRECATIONS DETECTED IN CODE:\n"
            for dep in deprecations:
                deprecations_text += f"• {dep['resource']}"
                if dep.get('field'):
                    deprecations_text += f".{dep['field']}"
                deprecations_text += f": {dep['message']}\n  (in {dep['file']})\n"
        
        # Build the prompt
        pr_reference_instruction = ""
        if pr_number and pr_url:
            pr_reference_instruction = f"5. Include the PR reference at the end: ([#{pr_number}]({pr_url}))"
        else:
            pr_reference_instruction = "5. Do NOT include any PR reference"
        
        prompt = f"""You are a technical writer for a Terraform provider. Generate a changelog entry in the style of the HashiCorp AWS Terraform provider.

COMMITS TO ANALYZE:
{commits_text}
{deprecations_text}

COMMIT TYPE PREFIXES TO RECOGNIZE:
- `feat:` - New resources, data sources, or major functionality
- `add:` - Adding new capabilities to existing resources
- `fix:` / `patch:` - Bug fixes and corrections
- `deprecate:` - Deprecating existing functionality
- `breaking:` - Breaking changes
- `enhance:` / `improve:` / `update:` - Enhancements and improvements
- `docs:` / `example:` - Documentation changes
- `refactor:` / `test:` / `chore:` / `ci:` - Internal changes (usually skip these)

REQUIREMENTS:
1. Categorize changes into these sections: {', '.join(self.config['categories'])}
2. Write clear, user-focused descriptions (not just commit messages)
3. Use commit type prefixes to guide categorization:
   - `feat:` / `add:` → FEATURES or ENHANCEMENTS (depending on scope)
   - `fix:` / `patch:` → BUG FIXES
   - `deprecate:` → DEPRECATIONS
   - `breaking:` → BREAKING CHANGES
   - `enhance:` / `improve:` / `update:` → ENHANCEMENTS
   - `docs:` / `example:` → DOCUMENTATION
   - Skip: `refactor:`, `test:`, `chore:`, `ci:` (unless significant user impact)
4. ONLY use resource notation for ACTUAL Terraform resources/data sources (files in rafay/ or internal/ that define resources):
   - For actual new resources: "* **New Resource:** `rafay_resource_name`"
   - For actual new data sources: "* **New Data Source:** `rafay_data_source_name`"
   - For changes to existing resources: "* resource/rafay_resource_name: Description of change"
{pr_reference_instruction}
6. For changes to tooling, automation, documentation, policies, or testing, use general descriptive format:
   - "* Implement automated changelog generation system"
   - "* Add deprecation policy documentation"
   - "* Improve testing framework for integration tests"
   - "* Update dependency versions for security patches"
7. Group related changes together intelligently
8. Prioritize significant changes, skip trivial ones (typos, minor refactoring, code comments)
9. Use present tense ("Add" not "Added", "Fix" not "Fixed")
10. If the change is not significant, skip it entirely
11. Do not include emojis in the changelog entries
12. Do NOT create fake resource names for non-resource changes (e.g., changelog scripts, docs, policies are NOT resources)

CATEGORIZATION RULES:

**BREAKING CHANGES** - Only for changes that break existing user configurations:
- Removing or renaming resources (e.g., "Remove rafay_cluster resource")
- Removing or renaming resource arguments/attributes
- Changing required vs optional status of fields
- Changing default values that affect existing deployments
- NOT for: removing comments, renaming variables, refactoring code, internal changes

**FEATURES** - New functionality that users can adopt:
- New resources (rafay_*)
- New data sources
- New optional arguments that add capabilities
- Major new functionality

**ENHANCEMENTS** - Improvements to existing functionality:
- Performance improvements
- Better error messages
- Additional validation
- Improved documentation
- Support for new cloud provider features

**BUG FIXES** - Corrections to incorrect behavior:
- Fixes for crashes, errors, or incorrect results
- Corrections to state management issues
- Fixes for import/export problems

**DEPRECATIONS** - Advance notice of future breaking changes:
- Deprecated resources, arguments, or values
- Include migration path in description

**DOCUMENTATION** - Documentation-only changes:
- Only include if significant (new guides, major rewrites)
- Skip minor typo fixes or formatting changes

Generate ONLY the changelog entries (bullet points), grouped by category. Do not include section headers, just the categorized bullet points."""

        try:
            response = self.client.chat.completions.create(
                model=self.config['ai_model'],
                max_tokens=2000,
                temperature=0.3,
                messages=[{
                    "role": "system",
                    "content": "You are a technical writer specializing in Terraform provider documentation. You generate professional changelog entries following HashiCorp AWS provider standards."
                }, {
                    "role": "user",
                    "content": prompt
                }]
            )
            
            content = response.choices[0].message.content
            
            # Post-process: Add PR reference if not already present and if valid
            # Only add PR reference if both pr_number and pr_url are provided and valid
            if pr_number and pr_url and pr_number != "None" and pr_url != "None":
                lines = content.split('\n')
                processed_lines = []
                for line in lines:
                    if line.strip().startswith('*') and f'#{pr_number}' not in line:
                        # Add PR reference before the last parenthesis or at the end
                        line = line.rstrip()
                        if line.endswith(')'):
                            line = line[:-1] + f" ([#{pr_number}]({pr_url}))"
                        else:
                            line = line + f" ([#{pr_number}]({pr_url}))"
                    processed_lines.append(line)
                content = '\n'.join(processed_lines)
            
            return content
        
        except Exception as e:
            print(f"Error generating changelog with AI: {e}")
            # Fallback: basic changelog from commit subjects
            return self._generate_fallback_changelog(commits, pr_number, pr_url)
    
    def _generate_fallback_changelog(self, commits: List[Dict], 
                                     pr_number: Optional[int],
                                     pr_url: Optional[str]) -> str:
        """Generate basic changelog without AI (fallback)."""
        entries = []
        for commit in commits:
            subject = commit['subject']
            # Only include PR reference if both are valid and not "None"
            ref = ""
            if pr_number and pr_url and pr_number != "None" and pr_url != "None":
                ref = f" ([#{pr_number}]({pr_url}))"
            entries.append(f"* {subject}{ref}")
        return '\n'.join(entries)
    
    def categorize_entries(self, content: str) -> Dict[str, List[str]]:
        """Parse AI-generated content and categorize entries."""
        categorized = {cat: [] for cat in self.config['categories']}
        
        current_category = None
        lines = content.split('\n')
        
        for line in lines:
            line = line.strip()
            if not line:
                continue
            
            # Check if line is a category header
            upper_line = line.upper().rstrip(':')
            if upper_line in self.config['categories']:
                current_category = upper_line
                continue
            
            # Check if line starts with bullet point
            if line.startswith('*') or line.startswith('-'):
                # Try to infer category from content
                line_lower = line.lower()
                
                if not current_category:
                    if any(kw in line_lower for kw in ['breaking', 'removed', 'renamed']):
                        current_category = 'BREAKING CHANGES'
                    elif any(kw in line_lower for kw in ['new resource', 'new data source']):
                        current_category = 'FEATURES'
                    elif any(kw in line_lower for kw in ['deprecat']):
                        current_category = 'DEPRECATIONS'
                    elif any(kw in line_lower for kw in ['fix', 'correct', 'resolve']):
                        current_category = 'BUG FIXES'
                    elif any(kw in line_lower for kw in ['add', 'improve', 'enhance', 'update']):
                        current_category = 'ENHANCEMENTS'
                    elif any(kw in line_lower for kw in ['doc']):
                        current_category = 'DOCUMENTATION'
                    else:
                        current_category = 'ENHANCEMENTS'
                
                categorized[current_category].append(line)
            elif current_category:
                # Non-bullet line but we have a category - include it as continuation
                # This handles multi-line entries or important context from AI
                categorized[current_category].append(line)
        
        return categorized
    
    def write_pr_changelog_fragment(self, entries: Dict[str, List[str]], 
                                    pr_number: int,
                                    changelog_dir: str = ".changelog") -> bool:
        """Write PR-specific changelog fragment to .changelog/ directory."""
        try:
            # Create .changelog directory if it doesn't exist
            os.makedirs(changelog_dir, exist_ok=True)
            
            # Build changelog fragment content
            fragment_content = ""
            for category in self.config['categories']:
                if entries.get(category):
                    fragment_content += f"### {category}\n\n"
                    for entry in entries[category]:
                        fragment_content += f"{entry}\n"
                    fragment_content += "\n"
            
            if not fragment_content.strip():
                print("No changelog entries to write.")
                return False
            
            # Write to PR-specific file
            fragment_path = os.path.join(changelog_dir, f"{pr_number}.txt")
            with open(fragment_path, 'w') as f:
                f.write(fragment_content)
            
            print(f"✓ Successfully wrote changelog fragment to {fragment_path}")
            return True
        
        except Exception as e:
            print(f"Error writing changelog fragment: {e}")
            return False
    
    def update_changelog_file(self, entries: Dict[str, List[str]], 
                             target_section: str = "Unreleased",
                             changelog_path: str = "CHANGELOG.md") -> bool:
        """Update CHANGELOG.md file with new entries."""
        try:
            # Read existing changelog
            if os.path.exists(changelog_path):
                with open(changelog_path, 'r') as f:
                    content = f.read()
            else:
                content = self._create_initial_changelog()
            
            # Find the target section
            if target_section == "Unreleased":
                section_pattern = r'(## Unreleased\s*\n)'
            else:
                section_pattern = f'(## {re.escape(target_section)}\\s*\\n)'
            
            # Build new content for the section
            new_section_content = ""
            for category in self.config['categories']:
                if entries.get(category):
                    new_section_content += f"\n### {category}\n\n"
                    for entry in entries[category]:
                        new_section_content += f"{entry}\n"
            
            if not new_section_content.strip():
                print("No changelog entries to add.")
                return False
            
            # Insert new content after section header
            match = re.search(section_pattern, content)
            if match:
                insert_pos = match.end()
                content = content[:insert_pos] + new_section_content + content[insert_pos:]
            else:
                # Section not found, append at the beginning after header
                header_end = content.find('\n## ')
                if header_end == -1:
                    content += f"\n## {target_section}\n{new_section_content}"
                else:
                    content = content[:header_end] + f"\n\n## {target_section}\n{new_section_content}" + content[header_end:]
            
            # Write updated changelog
            with open(changelog_path, 'w') as f:
                f.write(content)
            
            print(f"✓ Successfully updated {changelog_path}")
            return True
        
        except Exception as e:
            print(f"Error updating changelog file: {e}")
            return False
    
    def _create_initial_changelog(self) -> str:
        """Create initial CHANGELOG.md structure."""
        return f"""# Changelog

All notable changes to the Rafay Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## Unreleased

### BREAKING CHANGES

### FEATURES

### ENHANCEMENTS

### BUG FIXES

### DEPRECATIONS

### DOCUMENTATION

---
"""


def main():
    parser = argparse.ArgumentParser(description='Generate changelog entries using AI')
    parser.add_argument('--pr-number', type=int, help='Pull request number')
    parser.add_argument('--pr-url', type=str, help='Pull request URL')
    parser.add_argument('--base-ref', default='origin/master', help='Base git reference')
    parser.add_argument('--head-ref', default='HEAD', help='Head git reference')
    parser.add_argument('--target-section', default='Unreleased', help='Target changelog section')
    parser.add_argument('--deprecations-file', help='Path to deprecations JSON file')
    parser.add_argument('--changelog-path', default='CHANGELOG.md', help='Path to CHANGELOG.md')
    parser.add_argument('--config', default='.github/changelog-config.json', help='Config file path')
    parser.add_argument('--dry-run', action='store_true', help='Print output without updating file')
    
    args = parser.parse_args()
    
    # Debug: Print received arguments
    print("=" * 50)
    print("DEBUG: Received Arguments")
    print("=" * 50)
    print(f"PR Number: {args.pr_number}")
    print(f"PR URL: {args.pr_url}")
    print(f"Base Ref: {args.base_ref}")
    print(f"Head Ref: {args.head_ref}")
    print(f"Target Section: {args.target_section}")
    print(f"Deprecations File: {args.deprecations_file}")
    print("=" * 50)

    # Get API key from environment
    api_key = os.getenv('OPENAI_API_KEY')
    if not api_key:
        print("Error: OPENAI_API_KEY environment variable not set")
        sys.exit(1)
    
    # Initialize generator
    generator = ChangelogGenerator(api_key, args.config)
    
    # Get commits
    print(f"Fetching commits from {args.base_ref}...{args.head_ref}")
    commits = generator.get_commits_from_pr(
        pr_number=args.pr_number,
        base_ref=args.base_ref,
        head_ref=args.head_ref
    )
    
    if not commits:
        print("No commits found.")
        sys.exit(0)
    
    print(f"Found {len(commits)} commit(s)")
    
    # Load deprecations if provided
    deprecations = []
    if args.deprecations_file and os.path.exists(args.deprecations_file):
        with open(args.deprecations_file, 'r') as f:
            dep_data = json.load(f)
            deprecations = dep_data.get('deprecations', [])
        print(f"Loaded {len(deprecations)} deprecation(s)")
    
    # Generate changelog content
    print("Generating changelog with AI...")
    content = generator.generate_changelog_content(
        commits, 
        deprecations,
        pr_number=args.pr_number,
        pr_url=args.pr_url
    )
    
    # Categorize entries
    categorized_entries = generator.categorize_entries(content)
    
    # Print or update file
    if args.dry_run:
        print("\n=== Generated Changelog Entries ===\n")
        for category, entries in categorized_entries.items():
            if entries:
                print(f"### {category}\n")
                for entry in entries:
                    print(entry)
                print()
    else:
        # If PR number is provided, write both fragment and update main changelog
        success = True
        
        if args.pr_number:
            # Write PR-specific changelog fragment to .changelog/{PR_NUMBER}.txt
            fragment_success = generator.write_pr_changelog_fragment(
                categorized_entries,
                pr_number=args.pr_number
            )
            success = success and fragment_success
        
        # Always update the main CHANGELOG.md
        changelog_success = generator.update_changelog_file(
            categorized_entries,
            target_section=args.target_section,
            changelog_path=args.changelog_path
        )
        success = success and changelog_success
        
        sys.exit(0 if success else 1)


if __name__ == '__main__':
    main()

