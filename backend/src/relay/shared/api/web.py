from pathlib import Path

from fastapi import APIRouter, HTTPException
from fastapi.responses import FileResponse, Response


def web_router(dist_dir: Path) -> APIRouter:
    """Serves the exported Expo web build, falling back to index.html for client-side routes.

    Resolution order for /some/path: the file itself, some/path.html, some/path/index.html,
    then the root index.html. Anything under /api that no API route matched stays a 404.
    """
    root = dist_dir.resolve()
    router = APIRouter(include_in_schema=False)

    def find(relative: str) -> Path | None:
        candidates = (root / relative, root / f"{relative}.html", root / relative / "index.html")
        for candidate in candidates:
            resolved = candidate.resolve()
            if resolved.is_relative_to(root) and resolved.is_file():
                return resolved
        return None

    @router.get("/{path:path}")
    async def serve(path: str) -> Response:
        if path == "api" or path.startswith("api/"):
            raise HTTPException(status_code=404)
        found = find(path.strip("/")) if path else None
        return FileResponse(found or root / "index.html")

    return router
