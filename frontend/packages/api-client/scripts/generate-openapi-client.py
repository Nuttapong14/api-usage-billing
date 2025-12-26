#!/usr/bin/env python3
import json
import re
from pathlib import Path
from typing import Any, Dict, List

import yaml

ROOT = Path(__file__).resolve().parents[4]
CONTRACTS_DIR = ROOT / "specs" / "001-api-usage-billing" / "contracts"
OUTPUT_DIR = ROOT / "frontend" / "packages" / "api-client" / "src" / "generated"


def pascal_case(value: str) -> str:
    parts = re.split(r"[^a-zA-Z0-9]", value)
    return "".join(part.capitalize() for part in parts if part)


def camel_case(value: str) -> str:
    pascal = pascal_case(value)
    return pascal[:1].lower() + pascal[1:] if pascal else ""


def ts_prop_name(name: str) -> str:
    if re.match(r"^[A-Za-z_][A-Za-z0-9_]*$", name):
        return name
    return json.dumps(name)


def indent_lines(lines: List[str], spaces: int) -> List[str]:
    pad = " " * spaces
    return [pad + line if line else line for line in lines]


def schema_to_ts(schema: Dict[str, Any], namespace: str, indent: int = 0) -> str:
    if not schema:
        return "unknown"
    if "$ref" in schema:
        ref_name = schema["$ref"].split("/")[-1]
        return f"{namespace}.{ref_name}"
    if "oneOf" in schema:
        return " | ".join(schema_to_ts(item, namespace, indent) for item in schema["oneOf"])
    if "anyOf" in schema:
        return " | ".join(schema_to_ts(item, namespace, indent) for item in schema["anyOf"])
    if "allOf" in schema:
        return " & ".join(schema_to_ts(item, namespace, indent) for item in schema["allOf"])
    if "enum" in schema:
        return " | ".join(json.dumps(value) for value in schema["enum"])

    schema_type = schema.get("type")
    if schema_type == "array":
        items_ts = schema_to_ts(schema.get("items", {}), namespace, indent)
        if " | " in items_ts or " & " in items_ts or "\n" in items_ts or items_ts.startswith("{"):
            items_ts = f"({items_ts})"
        return f"{items_ts}[]"
    if schema_type == "object" or "properties" in schema:
        props = schema.get("properties", {})
        required = set(schema.get("required", []))
        additional = schema.get("additionalProperties")
        if isinstance(additional, bool):
            if additional:
                return "Record<string, unknown>"
            additional = None

        if not props and additional:
            return f"Record<string, {schema_to_ts(additional, namespace, indent)}>"

        lines = ["{"]
        for prop_name, prop_schema in props.items():
            optional = "" if prop_name in required else "?"
            ts_name = ts_prop_name(prop_name)
            ts_type = schema_to_ts(prop_schema, namespace, indent + 2)
            lines.append(f"  {ts_name}{optional}: {ts_type};")
        if additional:
            lines.append(f"  [key: string]: {schema_to_ts(additional, namespace, indent + 2)};")
        lines.append("}")
        return "\n".join(indent_lines(lines, indent))
    if schema_type == "integer" or schema_type == "number":
        return "number"
    if schema_type == "boolean":
        return "boolean"
    if schema_type == "string":
        return "string"

    return "unknown"


def first_json_schema(responses: Dict[str, Any]) -> Dict[str, Any]:
    for status, payload in responses.items():
        if not str(status).startswith("2"):
            continue
        content = payload.get("content", {})
        json_schema = content.get("application/json", {}).get("schema")
        if json_schema:
            return json_schema
    return {}


def load_spec(path: Path) -> Dict[str, Any]:
    with path.open("r", encoding="utf-8") as handle:
        return yaml.safe_load(handle)


def build_param_type(params: List[Dict[str, Any]], namespace: str, name: str) -> str:
    lines = [f"export interface {name} {{"]
    for param in params:
        schema = param.get("schema", {})
        ts_type = schema_to_ts(schema, namespace, 2)
        required = param.get("required", False)
        prop_name = ts_prop_name(param.get("name", "param"))
        optional = "" if required else "?"
        lines.append(f"  {prop_name}{optional}: {ts_type};")
    lines.append("}")
    return "\n".join(lines)


def generate_spec(spec_path: Path) -> None:
    spec = load_spec(spec_path)
    spec_name = pascal_case(spec_path.stem)
    client_name = camel_case(spec_path.stem)

    parameter_components = spec.get("components", {}).get("parameters", {})

    namespace_lines: List[str] = []

    components = spec.get("components", {}).get("schemas", {})
    for schema_name, schema in components.items():
        ts_type = schema_to_ts(schema, spec_name, 0)
        if ts_type.startswith("{"):
            namespace_lines.append(f"export interface {schema_name} {ts_type}")
        else:
            namespace_lines.append(f"export type {schema_name} = {ts_type};")

    path_params_defs: List[str] = []
    query_params_defs: List[str] = []
    request_param_defs: List[str] = []

    client_lines: List[str] = [f"export const {client_name}Client = {{"]

    paths = spec.get("paths", {})
    for path, operations in paths.items():
        for method, operation in operations.items():
            if method.lower() not in {"get", "post", "put", "patch", "delete"}:
                continue
            operation_id = operation.get("operationId")
            if not operation_id:
                operation_id = camel_case(f"{method}-{path}")

            op_name = pascal_case(operation_id)
            path_params: List[Dict[str, Any]] = []
            query_params: List[Dict[str, Any]] = []
            query_required = False

            combined_params = []
            combined_params.extend(operations.get("parameters", []))
            combined_params.extend(operation.get("parameters", []))

            for param in combined_params:
                if "$ref" in param:
                    ref_name = param["$ref"].split("/")[-1]
                    param = parameter_components.get(ref_name, param)
                if param.get("in") == "path":
                    path_params.append(param)
                if param.get("in") == "query":
                    query_params.append(param)
                    if param.get("required"):
                        query_required = True

            path_type_name = ""
            query_type_name = ""
            if path_params:
                path_type_name = f"{op_name}PathParams"
                path_params_defs.append(build_param_type(path_params, spec_name, path_type_name))
            if query_params:
                query_type_name = f"{op_name}QueryParams"
                query_params_defs.append(build_param_type(query_params, spec_name, query_type_name))

            body_schema = None
            body_required = False
            request_body = operation.get("requestBody")
            if request_body:
                body_required = request_body.get("required", False)
                content = request_body.get("content", {})
                body_schema = content.get("application/json", {}).get("schema")

            response_schema = first_json_schema(operation.get("responses", {}))
            response_type = schema_to_ts(response_schema, spec_name, 0) if response_schema else "unknown"

            params_type_name = f"{op_name}Params"
            param_fields: List[str] = []
            if path_type_name:
                param_fields.append(f"  path: {spec_name}.{path_type_name};")
            if query_type_name:
                param_fields.append(f"  query?: {spec_name}.{query_type_name};")
            if body_schema:
                body_type = schema_to_ts(body_schema, spec_name, 0)
                body_optional = "" if body_required else "?"
                param_fields.append(f"  body{body_optional}: {body_type};")

            if param_fields:
                request_param_defs.append(
                    "\n".join([f"export interface {params_type_name} {{"] + param_fields + ["}"])
                )
            params_required = bool(path_params) or body_required or query_required
            if param_fields:
                default_suffix = "" if params_required else " = {}"
                params_arg = f"params: {spec_name}.{params_type_name}{default_suffix}"
            else:
                params_arg = "params: Record<string, never> = {}"

            path_literal = f"v1{path}"
            client_lines.append(
                f"  async {operation_id}({params_arg}, client: APIClient = getDefaultClient()): "
                f"Promise<APIResponse<{response_type}>> {{"
            )
            if path_params:
                client_lines.append(f"    const url = interpolatePath('{path_literal}', params.path);")
            else:
                client_lines.append(f"    const url = '{path_literal}';")
            if query_params:
                client_lines.append("    const query = params.query as Record<string, string | number | boolean> | undefined;")
            else:
                client_lines.append("    const query = undefined;")
            method_lower = method.lower()
            if method_lower in {"post", "put", "patch"}:
                if body_schema:
                    client_lines.append(f"    return client.{method_lower}< {response_type} >(url, params.body, {{ params: query }});")
                else:
                    client_lines.append(f"    return client.{method_lower}< {response_type} >(url, undefined, {{ params: query }});")
            else:
                client_lines.append(f"    return client.{method_lower}< {response_type} >(url, {{ params: query }});")
            client_lines.append("  },")

    client_lines.append("};")

    output_lines: List[str] = [
        "import { APIClient, getDefaultClient } from '../client';",
        "import { APIResponse } from '../types';",
        "import { interpolatePath } from './runtime';",
        "",
        f"export namespace {spec_name} {{",
    ]

    for block in namespace_lines + path_params_defs + query_params_defs + request_param_defs:
        output_lines.extend(indent_lines(block.split("\n"), 2))
        output_lines.append("")

    output_lines.append("}")
    output_lines.append("")
    output_lines.extend(client_lines)
    output_lines.append("")

    OUTPUT_DIR.mkdir(parents=True, exist_ok=True)
    out_path = OUTPUT_DIR / f"{spec_path.stem}.ts"
    out_path.write_text("\n".join(output_lines).rstrip() + "\n", encoding="utf-8")


def main() -> None:
    if not CONTRACTS_DIR.exists():
        raise SystemExit(f"Contracts directory not found: {CONTRACTS_DIR}")

    specs = sorted(CONTRACTS_DIR.glob("*.yaml"))
    if not specs:
        raise SystemExit("No contract files found.")

    for spec in specs:
        generate_spec(spec)

    index_lines = ["export * from './runtime';"]
    for spec in specs:
        index_lines.append(f"export * from './{spec.stem}';")

    (OUTPUT_DIR / "index.ts").write_text("\n".join(index_lines).rstrip() + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
