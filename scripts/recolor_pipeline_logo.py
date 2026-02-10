#!/usr/bin/env python3
"""Recolore le logo Pipelines : flèches en cyan, fond gris -> transparent.
   À lancer depuis la racine du projet : python scripts/recolor_pipeline_logo.py
   Prérequis : pip install Pillow
"""
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("Installez Pillow : pip install Pillow")
    raise SystemExit(1)

SCRIPT_DIR = Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parent
SRC = ROOT / "frontend" / "src" / "assets" / "pipeline_logo.png"
OUT = SRC

CYAN = (0, 229, 255)  # #00E5FF


def is_gray_background(r, g, b, a):
    """Fond gris clair (carré derrière le logo)."""
    if a < 100:
        return True
    # Gris : R,G,B proches, valeur moyenne élevée
    avg = (r + g + b) / 3
    return 160 <= avg <= 255 and abs(r - g) < 40 and abs(g - b) < 40 and abs(r - b) < 40


def is_drawing(r, g, b, a):
    """Noir, blanc ou contour (flèches / symbole)."""
    if a < 25:
        return False
    # Noir / très sombre
    if r < 80 and g < 80 and b < 80:
        return True
    # Blanc / clair (contour)
    if r > 200 and g > 200 and b > 200:
        return True
    # Gris moyen (anti-aliasing du dessin)
    avg = (r + g + b) / 3
    if 40 <= avg <= 200:
        return True
    return False


def main():
    if not SRC.exists():
        print(f"Fichier introuvable: {SRC}")
        return 1
    img = Image.open(SRC).convert("RGBA")
    w, h = img.size
    data = list(img.getdata())
    new_data = []
    for (r, g, b, a) in data:
        if is_gray_background(r, g, b, a):
            new_data.append((0, 0, 0, 0))
        elif is_drawing(r, g, b, a):
            # Garder l'alpha pour les bords lisses
            new_a = max(a, 180) if (r > 200 or a > 150) else a
            new_data.append((*CYAN, new_a))
        else:
            new_data.append((0, 0, 0, 0))
    img.putdata(new_data)
    img.save(OUT, "PNG")
    print(f"Logo Pipelines recoloré (cyan, fond transparent) : {OUT}")
    return 0


if __name__ == "__main__":
    exit(main())
