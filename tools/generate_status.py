#!/usr/bin/env python3
"""Generate a status.yaml-style file from configPC.yaml and configPJ.yaml.

For every zone (top-level key) found in configPC.yaml a zone entry is created
with a "pc_1" status field. If the same zone key also exists in configPJ.yaml,
every pj_N key found there is added to the zone's status as well.

Usage:
    python generate_status.py [--pc configPC.yaml] [--pj configPJ.yaml]
                               [--out status_generated.yaml]
                               [--pc-value 521] [--pj-value 520]

python tools/generate_status.py --pc configs/configPC.yaml --pj configs/configPJ.yaml --out configs/status_generated.yaml
"""

import argparse
from pathlib import Path

import json
import yaml

DEFAULT_PC_VALUE = 521
DEFAULT_PJ_VALUE = 520


class IndentedDumper(yaml.Dumper):
    """Indents list items under their parent key, matching status.yaml's style."""

    def increase_indent(self, flow=False, indentless=False):
        return super().increase_indent(flow, False)


def load_yaml(path: Path) -> dict:
    with path.open("r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    if not isinstance(data, dict):
        raise ValueError(f"{path} does not contain a top-level mapping")
    return data


def build_zones(pc: dict, pj: dict, pc_value: int, pj_value: int) -> list:
    zones = []
    for zone_id, pc_info in pc.items():
        status = {}

        # configPC.yaml zones always have a single PC, exposed as pc_1
        if isinstance(pc_info, dict) and pc_info:
            status["pc_1"] = pc_value

        pj_info = pj.get(zone_id)
        if isinstance(pj_info, dict):
            for pj_key in sorted(pj_info.keys()):
                status[pj_key] = pj_value

        zones.append({"id": zone_id, "status": status})
    return zones


def main() -> None:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--pc", default="configs/configPC.yaml", type=Path)
    parser.add_argument("--pj", default="configs/configPJ.yaml", type=Path)
    parser.add_argument("--out", default="configs/status_generated.yaml", type=Path)
    parser.add_argument("--pc-value", default=DEFAULT_PC_VALUE, type=int)
    parser.add_argument("--pj-value", default=DEFAULT_PJ_VALUE, type=int)
    args = parser.parse_args()

    pc = load_yaml(args.pc)
    pj = load_yaml(args.pj)

    zones = build_zones(pc, pj, args.pc_value, args.pj_value)

    args.out.parent.mkdir(parents=True, exist_ok=True)
    with args.out.open("w", encoding="utf-8") as f:
        yaml.dump(
            {"zones": zones},
            f,
            Dumper=IndentedDumper,
            allow_unicode=True,
            sort_keys=False,
            default_flow_style=False,
        )

    with open(args.out, "r", encoding="utf-8") as yaml_in, open(
        args.out.with_suffix(".json"), "w", encoding="utf-8"
    ) as json_out:
        yaml_object = yaml.safe_load(yaml_in)  # yaml_object will be a list or a dict
        json.dump(yaml_object, json_out, sort_keys=False, indent=2, ensure_ascii=True)

    print(f"Written {len(zones)} zones to {args.out}")


if __name__ == "__main__":
    main()
