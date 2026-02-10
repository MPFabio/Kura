#!/usr/bin/env python3
"""Recolore le logo Kubernetes : bleu -> cyan, blanc -> couleur du fond.
   À lancer depuis la racine du projet : python scripts/recolor_k8s_logo.py
   Prérequis : pip install Pillow
"""
from pathlib import Path

try:
    from PIL import Image
except ImportError:
    print("Installez Pillow : pip install Pillow")
    raise SystemExit(1)

# Chemins (depuis la racine du repo)
SCRIPT_DIR = Path(__file__).resolve().parent
ROOT = SCRIPT_DIR.parent
SRC = ROOT / "frontend" / "src" / "assets" / "k8s_logo.png"
OUT = SRC

# Couleurs cibles (theme KURA)
CYAN = (0, 229, 255)           # #00E5FF - hexagone
BG = (44, 47, 63)              # #2c2f3f - fond (roue = même couleur que le fond)

# Seuils pour détecter bleu K8s (~#326CE5) et blanc (avec anti-aliasing)
def is_white(r, g, b, a):
    return a > 120 and r > 200 and g > 200 and b > 200

def is_blue(r, g, b, a):
    return a > 80 and b > max(r, g) and g > 60 and r < 120

def main():
    if not SRC.exists():
        print(f"Fichier introuvable: {SRC}")
        return 1
    img = Image.open(SRC).convert("RGBA")
    w, h = img.size
    data = list(img.getdata())
    new_data = []
    for (r, g, b, a) in data:
        if a < 10:
            new_data.append((r, g, b, a))
        elif is_white(r, g, b, a):
            new_data.append((*BG, a))
        elif is_blue(r, g, b, a):
            new_data.append((*CYAN, a))
        else:
            new_data.append((r, g, b, a))
    img.putdata(new_data)
    img.save(OUT, "PNG")
    print(f"Logo recoloré enregistré: {OUT}")
    return 0

if __name__ == "__main__":
    exit(main())
