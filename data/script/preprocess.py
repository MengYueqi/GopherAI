import argparse
import json
import os
import random
from typing import Dict, Iterable, List, Optional


def _reservoir_sample_jsonl(path: str, k: int, seed: Optional[int]) -> List[Dict]:
    rng = random.Random(seed)
    sample: List[Dict] = []
    with open(path, "r", encoding="utf-8") as f:
        for i, line in enumerate(f):
            line = line.strip()
            if not line:
                continue
            item = json.loads(line)
            if len(sample) < k:
                sample.append(item)
                continue
            j = rng.randint(0, i)
            if j < k:
                sample[j] = item
    return sample


def _load_json_array(path: str) -> List[Dict]:
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def _is_jsonl(path: str) -> bool:
    if path.endswith(".jsonl"):
        return True
    with open(path, "r", encoding="utf-8") as f:
        for line in f:
            stripped = line.strip()
            if not stripped:
                continue
            return stripped.startswith("{")
    return False


def _compose_content(item: Dict) -> Dict:
    question = str(item.get("question", "")).strip()
    answer = str(item.get("answer", "")).strip()
    content = f"{question}\n{answer}".strip()
    return {"id": item.get("id"), "content": content}


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Sample medical data and build id+content JSON."
    )
    parser.add_argument(
        "--input",
        default=os.path.join("data", "all_medical_data.jsonl"),
        help="Input JSONL or JSON array file.",
    )
    parser.add_argument(
        "--output",
        default=os.path.join("data", "all_medical_sample.json"),
        help="Output JSON file.",
    )
    parser.add_argument(
        "--count",
        type=int,
        default=1000,
        help="Number of items to sample.",
    )
    parser.add_argument("--seed", type=int, default=None, help="Random seed.")
    args = parser.parse_args()

    if _is_jsonl(args.input):
        items = _reservoir_sample_jsonl(args.input, args.count, args.seed)
    else:
        items = _load_json_array(args.input)
        if args.seed is not None:
            random.Random(args.seed).shuffle(items)
        else:
            random.shuffle(items)
        items = items[: args.count]

    output = [_compose_content(item) for item in items]

    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(output, f, ensure_ascii=False, indent=2)


if __name__ == "__main__":
    main()
