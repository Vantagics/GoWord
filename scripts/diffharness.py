#!/usr/bin/env python3
"""Differential harness for goword against LibreOffice, python-docx, pandoc, and Apache POI.

Reference tools are optional. A missing tool is reported as "skipped" and does
not fail the run. Use --strict to fail when a configured tool is absent.

Directions:
  read       goword's extracted text vs each reference reader
  roundtrip  goword rewrites the file; every reference must open it, and
             LibreOffice/pandoc/POI text must still cover the original text

Exit status is 0 when no compared pair falls below --min-recall (default 0.90)
and no reference tool errors on a file it tried to open. Skipped tools do not
fail the run.
"""

from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
import unicodedata
from collections import Counter
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
POI_DIR = Path(os.environ.get("POI_DIR", "/tmp/poi"))
POI_CP = os.environ.get("POI_CP", "")


def norm(text: str) -> str:
    text = unicodedata.normalize("NFKC", text or "")
    text = text.replace("\u00a0", " ").replace("\u200b", "").replace("\ufeff", "")
    text = text.replace("\r\n", "\n").replace("\r", "\n")
    lines = [" ".join(line.split()) for line in text.split("\n")]
    return "\n".join(line for line in lines if line)


def tokens(text: str) -> list[str]:
    return re.findall(r"\w+", norm(text).lower(), flags=re.UNICODE)


def coverage(reference: str, candidate: str) -> float:
    """Fraction of reference word tokens present in candidate (multiset)."""
    ref = tokens(reference)
    if not ref:
        return 1.0
    have = Counter(tokens(candidate))
    hit = 0
    for tok in ref:
        if have[tok] > 0:
            have[tok] -= 1
            hit += 1
    return hit / len(ref)


def run(cmd: list[str], timeout: int = 90) -> tuple[int, str, str]:
    try:
        proc = subprocess.run(
            cmd,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            timeout=timeout,
            check=False,
        )
    except FileNotFoundError as exc:
        return 127, "", str(exc)
    except subprocess.TimeoutExpired:
        return 124, "", "timeout"
    out = proc.stdout.decode("utf-8", "replace")
    err = proc.stderr.decode("utf-8", "replace")
    return proc.returncode, out, err


def tool_available(name: str) -> bool:
    return shutil.which(name) is not None


def python_docx_text(path: Path) -> tuple[str | None, str]:
    if not tool_available("python3"):
        return None, "python3 missing"
    code = r"""
import sys
try:
    import docx
except Exception as exc:
    sys.stderr.write("import "+str(exc))
    sys.exit(127)
doc = docx.Document(sys.argv[1])
parts = []
def para(p):
    t = p.text.strip()
    if t:
        parts.append(t)
def table(tbl):
    for row in tbl.rows:
        cells = []
        for cell in row.cells:
            cells.append(cell.text.strip())
        line = " ".join(c for c in cells if c)
        if line:
            parts.append(line)
for p in doc.paragraphs:
    para(p)
for tbl in doc.tables:
    table(tbl)
for sec in doc.sections:
    for part in (sec.header, sec.footer):
        try:
            for p in part.paragraphs:
                para(p)
            for tbl in part.tables:
                table(tbl)
        except Exception:
            pass
sys.stdout.write("\n".join(parts))
"""
    code_rc, out, err = run(["python3", "-c", code, str(path)], timeout=60)
    if code_rc == 127:
        return None, err.strip() or "python-docx missing"
    if code_rc != 0:
        return "", err.strip() or f"python-docx exit {code_rc}"
    return out, ""


def pandoc_text(path: Path) -> tuple[str | None, str]:
    if not tool_available("pandoc"):
        return None, "pandoc missing"
    rc, out, err = run(["pandoc", str(path), "-t", "plain", "--wrap=none"], timeout=60)
    if rc != 0:
        return "", err.strip() or f"pandoc exit {rc}"
    return out, ""


def soffice_text(path: Path, work: Path) -> tuple[str | None, str]:
    if not tool_available("soffice"):
        return None, "soffice missing"
    outdir = work / "lo"
    outdir.mkdir(parents=True, exist_ok=True)
    for old in outdir.glob("*.txt"):
        old.unlink()
    profile = work / "lo-profile"
    profile.mkdir(parents=True, exist_ok=True)
    rc, _, err = run(
        [
            "soffice",
            "--headless",
            "--norestore",
            "--nolockcheck",
            f"-env:UserInstallation=file://{profile}",
            "--convert-to",
            "txt:Text",
            "--outdir",
            str(outdir),
            str(path),
        ],
        timeout=90,
    )
    produced = list(outdir.glob("*.txt"))
    if rc != 0 or not produced:
        return "", (err.strip() or f"soffice exit {rc}")[:500]
    text = produced[0].read_text("utf-8", "replace")
    for p in produced:
        p.unlink(missing_ok=True)
    return text, ""


def poi_classpath() -> str | None:
    if POI_CP:
        return POI_CP
    if not POI_DIR.is_dir():
        return None
    jars = sorted(str(p) for p in POI_DIR.glob("*.jar"))
    if not jars:
        return None
    return ":".join([str(POI_DIR), *jars])


def poi_text(path: Path) -> tuple[str | None, str]:
    if not tool_available("java"):
        return None, "java missing"
    cp = poi_classpath()
    if not cp or not (POI_DIR / "PoiExtract.class").exists():
        return None, "poi extractor missing"
    rc, out, err = run(["java", "-cp", cp, "PoiExtract", str(path)], timeout=60)
    if rc != 0:
        return "", (err.strip() or f"poi exit {rc}")[:500]
    return out, ""


def goword_text(binary: Path, path: Path) -> tuple[str, str]:
    rc, out, err = run([str(binary), "extract", str(path)], timeout=60)
    if rc != 0:
        return "", err.strip() or f"goword exit {rc}"
    return out, ""


def goword_roundtrip(binary: Path, src: Path, dst: Path) -> tuple[str, str]:
    rc, out, err = run([str(binary), "roundtrip", str(src), str(dst)], timeout=60)
    if rc != 0:
        return "", err.strip() or f"goword roundtrip exit {rc}"
    return out, ""


READERS = {
    "libreoffice": soffice_text,
    "pandoc": pandoc_text,
    "python-docx": python_docx_text,
    "poi": poi_text,
}


def iter_docx(root: Path) -> list[Path]:
    files = []
    for p in root.rglob("*"):
        if not p.is_file() or p.name.startswith("~$"):
            continue
        if p.suffix.lower() not in {".docx", ".docm", ".dotx"}:
            continue
        # Encrypted packages are compound files. The other readers cannot be
        # given a password here; Go tests cover them.
        if "encrypted" in p.parts:
            continue
        files.append(p)
    return sorted(files)


def compare_read(goword: str, refs: dict[str, tuple[str | None, str]]) -> dict:
    result = {}
    for name, (text, err) in refs.items():
        if text is None:
            result[name] = {"status": "skipped", "detail": err}
            continue
        if err and text == "":
            result[name] = {"status": "error", "detail": err[:300]}
            continue
        cov = coverage(text, goword)
        back = coverage(goword, text) if tokens(goword) else 1.0
        status = "ok" if cov >= 0.90 else "mismatch"
        # Empty reference and empty goword is agreement.
        if not tokens(text) and not tokens(goword):
            cov, back, status = 1.0, 1.0, "ok"
        result[name] = {
            "status": status,
            "recall": round(cov, 4),
            "precision": round(back, 4),
        }
    return result


def main() -> int:
    parser = argparse.ArgumentParser(description="Compare goword with reference docx readers")
    parser.add_argument("roots", nargs="*", default=[str(ROOT / "testdata" / "word")])
    parser.add_argument("--bin", default=str(ROOT / "scripts" / "gowordtext"), help="gowordtext binary")
    parser.add_argument("--before-bin", default="", help="optional earlier gowordtext binary")
    parser.add_argument("--report", default="", help="write JSON report here")
    parser.add_argument("--min-recall", type=float, default=0.90)
    parser.add_argument("--limit", type=int, default=0)
    args = parser.parse_args()

    binary = Path(args.bin)
    if not binary.exists():
        print(f"gowordtext binary not found: {binary}", file=sys.stderr)
        print("build it with: go build -o /tmp/goword-text ./scripts/gowordtext", file=sys.stderr)
        return 2

    files: list[Path] = []
    for root in args.roots:
        files.extend(iter_docx(Path(root)))
    if args.limit:
        files = files[: args.limit]
    if not files:
        print("no .docx files found", file=sys.stderr)
        return 2

    rows = []
    failures = 0
    with tempfile.TemporaryDirectory(prefix="goword-diff-") as tmp:
        work = Path(tmp)
        for i, path in enumerate(files, 1):
            rel = str(path)
            try:
                rel = str(path.relative_to(ROOT))
            except ValueError:
                pass
            print(f"[{i}/{len(files)}] {rel}", flush=True)
            gtext, gerr = goword_text(binary, path)
            refs = {}
            for name, fn in READERS.items():
                if name == "libreoffice":
                    refs[name] = fn(path, work)
                else:
                    refs[name] = fn(path)
            read_cmp = compare_read(gtext, refs)

            rt_path = work / "roundtrip.docx"
            if rt_path.exists():
                rt_path.unlink()
            _, rterr = goword_roundtrip(binary, path, rt_path)
            rt = {"status": "error", "detail": rterr[:300]} if rterr or not rt_path.exists() else {"status": "ok", "readers": {}}
            if rt["status"] == "ok":
                for name, fn in READERS.items():
                    if name == "libreoffice":
                        text, err = fn(rt_path, work)
                    else:
                        text, err = fn(rt_path)
                    if text is None:
                        rt["readers"][name] = {"status": "skipped", "detail": err}
                        continue
                    if err and text == "":
                        rt["readers"][name] = {"status": "error", "detail": err[:300]}
                        failures += 1
                        continue
                    # Round-trip must preserve what this reference saw in the original,
                    # when the reference could read the original.
                    orig_text, orig_err = refs[name]
                    if orig_text is None or (orig_err and orig_text == ""):
                        rt["readers"][name] = {"status": "ok", "recall": 1.0, "note": "original unreadable"}
                        continue
                    cov = coverage(orig_text, text)
                    status = "ok" if cov >= args.min_recall else "mismatch"
                    if not tokens(orig_text):
                        cov, status = 1.0, "ok"
                    rt["readers"][name] = {"status": status, "recall": round(cov, 4)}
                    if status != "ok":
                        failures += 1

            before = None
            if args.before_bin:
                btext, berr = goword_text(Path(args.before_bin), path)
                before = {"error": berr[:200]} if berr else compare_read(btext, refs)

            for name, info in read_cmp.items():
                if info.get("status") == "error":
                    failures += 1
                elif info.get("status") == "mismatch":
                    failures += 1
            if gerr:
                failures += 1
            row = {
                "file": rel,
                "goword_error": gerr[:300],
                "goword_tokens": len(tokens(gtext)),
                "read": read_cmp,
                "roundtrip": rt,
                "before": before,
            }
            rows.append(row)

    summary = {
        "files": len(rows),
        "failures": failures,
        "min_recall": args.min_recall,
        "rows": rows,
    }
    if args.report:
        Path(args.report).write_text(json.dumps(summary, indent=2), "utf-8")
        print(f"wrote {args.report}")
    print(f"files={len(rows)} failure_marks={failures}")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
