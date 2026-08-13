#!/usr/bin/env python3
"""Gera os diagramas de documentação do desafio (user-flow.png e data-flow.png).

Uso:
    python3 docs/generate_diagrams.py

Requisito: Pillow (`pip install pillow`). Os PNGs já versionados em /docs foram
produzidos por este script — ele fica no repositório para que os diagramas
possam ser regerados quando o fluxo mudar.
"""

from __future__ import annotations

import os
from PIL import Image, ImageDraw, ImageFont

SCALE = 2  # desenha em 2x e reduz no final (antialiasing)

BG = (246, 245, 242)
INK = (35, 32, 28)
MUTED = (109, 103, 94)
LINE = (176, 169, 158)
WHITE = (255, 255, 255)

BRAND = (185, 139, 52)
BLUE = (63, 123, 201)
GREEN = (63, 160, 106)
RED = (179, 57, 47)
PURPLE = (124, 92, 168)

FONT_DIR = "/usr/share/fonts/truetype/dejavu"
REGULAR = os.path.join(FONT_DIR, "DejaVuSans.ttf")
BOLD = os.path.join(FONT_DIR, "DejaVuSans-Bold.ttf")
MONO = os.path.join(FONT_DIR, "DejaVuSansMono.ttf")


def font(path: str, size: int) -> ImageFont.FreeTypeFont:
    return ImageFont.truetype(path, size * SCALE)


class Canvas:
    def __init__(self, width: int, height: int):
        self.width, self.height = width, height
        self.image = Image.new("RGB", (width * SCALE, height * SCALE), BG)
        self.draw = ImageDraw.Draw(self.image)

    # --- primitivas ---

    def text(self, xy, content, *, f, fill=INK, anchor="mm"):
        x, y = xy
        self.draw.text((x * SCALE, y * SCALE), content, font=f, fill=fill, anchor=anchor)

    def box(self, rect, *, title=None, lines=(), accent=BRAND, fill=WHITE, mono_lines=(), radius=10):
        """Caixa arredondada com faixa colorida à esquerda."""
        x1, y1, x2, y2 = (v * SCALE for v in rect)
        self.draw.rounded_rectangle(
            (x1, y1, x2, y2), radius=radius * SCALE, fill=fill, outline=LINE, width=1 * SCALE
        )
        self.draw.rounded_rectangle(
            (x1, y1, x1 + 5 * SCALE, y2), radius=3 * SCALE, fill=accent
        )

        cx = (rect[0] + rect[2]) / 2
        blocks = []
        if title:
            blocks.append((title, font(BOLD, 14), INK, 20))
        for line in lines:
            blocks.append((line, font(REGULAR, 13), MUTED, 18))
        for line in mono_lines:
            blocks.append((line, font(MONO, 12), accent, 18))

        total = sum(height for *_, height in blocks)
        y = (rect[1] + rect[3]) / 2 - total / 2 + blocks[0][3] / 2 if blocks else 0
        for content, f, fill_color, height in blocks:
            self.text((cx, y), content, f=f, fill=fill_color)
            y += height

    def arrow(self, start, end, *, color=LINE, label=None, label_side="right", dashed=False, width=2):
        x1, y1 = (v * SCALE for v in start)
        x2, y2 = (v * SCALE for v in end)

        if dashed:
            self._dashed_line((x1, y1), (x2, y2), color, width)
        else:
            self.draw.line((x1, y1, x2, y2), fill=color, width=width * SCALE)

        # ponta da seta
        import math

        angle = math.atan2(y2 - y1, x2 - x1)
        size = 9 * SCALE
        self.draw.polygon(
            [
                (x2, y2),
                (x2 - size * math.cos(angle - 0.45), y2 - size * math.sin(angle - 0.45)),
                (x2 - size * math.cos(angle + 0.45), y2 - size * math.sin(angle + 0.45)),
            ],
            fill=color,
        )

        if label:
            mx, my = (start[0] + end[0]) / 2, (start[1] + end[1]) / 2
            offset = 10 if label_side == "right" else -10
            anchor = "lm" if label_side == "right" else "rm"
            if abs(end[0] - start[0]) > abs(end[1] - start[1]):  # seta horizontal
                self.text((mx, my - 12), label, f=font(REGULAR, 12), fill=MUTED, anchor="mm")
            else:
                self.text((mx + offset, my), label, f=font(REGULAR, 12), fill=MUTED, anchor=anchor)

    def _dashed_line(self, start, end, color, width):
        import math

        x1, y1 = start
        x2, y2 = end
        length = math.hypot(x2 - x1, y2 - y1)
        dash, gap = 8 * SCALE, 6 * SCALE
        steps = max(int(length // (dash + gap)), 1)
        for i in range(steps + 1):
            t1 = min((i * (dash + gap)) / length, 1)
            t2 = min((i * (dash + gap) + dash) / length, 1)
            self.draw.line(
                (x1 + (x2 - x1) * t1, y1 + (y2 - y1) * t1, x1 + (x2 - x1) * t2, y1 + (y2 - y1) * t2),
                fill=color,
                width=width * SCALE,
            )

    def elbow(self, points, *, color=LINE, label=None, width=2):
        """Liga uma sequência de pontos em ângulo reto, com seta no último trecho."""
        for i in range(len(points) - 2):
            a, b = points[i], points[i + 1]
            self.draw.line(
                (a[0] * SCALE, a[1] * SCALE, b[0] * SCALE, b[1] * SCALE), fill=color, width=width * SCALE
            )
        self.arrow(points[-2], points[-1], color=color, label=label, width=width)

    def save(self, path: str):
        final = self.image.resize((self.width, self.height), Image.LANCZOS)
        final.save(path, optimize=True)
        print(f"gerado: {path}")


# --------------------------------------------------------------------------
# user-flow.png
# --------------------------------------------------------------------------

def user_flow(path: str):
    c = Canvas(1600, 1160)

    c.text((800, 46), "User Flow — Mini Kanban de Tarefas", f=font(BOLD, 26))
    c.text(
        (800, 76),
        "Ações do usuário no quadro e a chamada de API correspondente",
        f=font(REGULAR, 14),
        fill=MUTED,
    )

    c.box((600, 108, 1000, 160), title="Usuário abre o Mini Kanban", accent=INK)
    c.arrow((800, 160), (800, 196))

    c.box(
        (560, 196, 1040, 262),
        title="Frontend carrega o quadro",
        mono_lines=("GET /api/tasks",),
        accent=BRAND,
    )

    # ramo de erro no carregamento
    c.box(
        (1120, 190, 1540, 268),
        title="Erro ao carregar",
        lines=("Faixa vermelha com a mensagem",),
        mono_lines=('botão "Tentar novamente"',),
        accent=RED,
    )
    c.arrow((1040, 214), (1120, 214), color=RED)
    c.text((1080, 236), "falha de rede", f=font(REGULAR, 12), fill=MUTED)
    c.elbow([(1330, 268), (1330, 300), (800, 300), (800, 306)], color=RED)

    c.box(
        (480, 306, 1120, 372),
        title="Quadro renderizado com três colunas fixas",
        lines=("A Fazer  •  Em Progresso  •  Concluídas",),
        accent=GREEN,
    )
    c.arrow((800, 372), (800, 408))

    c.box((640, 408, 960, 456), title="O que o usuário quer fazer?", accent=MUTED, fill=(240, 238, 234))

    lanes = [
        {
            "x": 230,
            "accent": BRAND,
            "steps": [
                ("Criar tarefa", ["Preenche o título (obrigatório)", "e a descrição (opcional)"], []),
                ("Envia o formulário", ["Título vazio é barrado antes", "de chamar a API"], []),
                ("API cria a tarefa", [], ["POST /api/tasks", "201 Created"]),
                ("Cartão surge em A Fazer", ["Campo de erro se a API recusar"], []),
            ],
        },
        {
            "x": 603,
            "accent": BLUE,
            "steps": [
                ("Editar tarefa", ['Clica em "Editar" no cartão'], []),
                ("Formulário inline", ["Altera título e/ou descrição", "Esc cancela"], []),
                ("API salva a alteração", [], ["PUT /api/tasks/{id}", "200 OK"]),
                ("Cartão atualizado", ["Rollback se a API recusar"], []),
            ],
        },
        {
            "x": 976,
            "accent": PURPLE,
            "steps": [
                ("Mover entre colunas", ["Arrasta o cartão (drag & drop)", "ou usa os botões ‹ ›"], []),
                ("UI move na hora", ["Atualização otimista com", "marcador da posição de destino"], []),
                ("API grava a nova coluna", [], ["PUT /api/tasks/{id}", "{status, position}"]),
                ("Coluna e ordem mantidas", ["Erro devolve o cartão ao lugar"], []),
            ],
        },
        {
            "x": 1349,
            "accent": RED,
            "steps": [
                ("Excluir tarefa", ['Clica em "Excluir" no cartão'], []),
                ("Confirmação no cartão", ['"Excluir esta tarefa?" Sim / Não'], []),
                ("API remove a tarefa", [], ["DELETE /api/tasks/{id}", "204 No Content"]),
                ("Cartão some do quadro", ["Erro restaura o cartão"], []),
            ],
        },
    ]

    lane_width = 340
    top = 520
    box_height = 92
    gap = 38

    for lane in lanes:
        x = lane["x"]
        c.arrow((800, 456), (x, top - 18), color=lane["accent"])

        y = top
        for index, (title, lines, mono_lines) in enumerate(lane["steps"]):
            c.box(
                (x - lane_width / 2, y, x + lane_width / 2, y + box_height),
                title=title,
                lines=lines,
                mono_lines=mono_lines,
                accent=lane["accent"],
            )
            if index < len(lane["steps"]) - 1:
                c.arrow((x, y + box_height), (x, y + box_height + gap - 4), color=lane["accent"])
            y += box_height + gap

    # As quatro trilhas descem até uma linha coletora e entram juntas na caixa final.
    lanes_bottom = top + 4 * (box_height + gap) - gap
    collector_y = lanes_bottom + 40
    bottom_y = collector_y + 30

    for lane in lanes:
        c.draw.line(
            (lane["x"] * SCALE, lanes_bottom * SCALE, lane["x"] * SCALE, collector_y * SCALE),
            fill=lane["accent"],
            width=2 * SCALE,
        )
    c.draw.line(
        (lanes[0]["x"] * SCALE, collector_y * SCALE, lanes[-1]["x"] * SCALE, collector_y * SCALE),
        fill=LINE,
        width=2 * SCALE,
    )
    c.arrow((800, collector_y), (800, bottom_y - 2))

    c.box(
        (520, bottom_y, 1080, bottom_y + 62),
        title="Quadro e API sincronizados",
        lines=("Estado em memória e, se habilitado, em data/tasks.json",),
        accent=GREEN,
    )

    c.save(path)


# --------------------------------------------------------------------------
# data-flow.png
# --------------------------------------------------------------------------

def data_flow(path: str):
    c = Canvas(1500, 980)

    c.text((750, 46), "Fluxo de Dados — Frontend, API e Persistência", f=font(BOLD, 26))
    c.text(
        (750, 76),
        "Do clique no navegador até o arquivo JSON, e o caminho de volta",
        f=font(REGULAR, 14),
        fill=MUTED,
    )

    # faixas
    c.draw.rounded_rectangle(
        (40 * SCALE, 120 * SCALE, 520 * SCALE, 900 * SCALE),
        radius=14 * SCALE,
        fill=(240, 238, 234),
        outline=LINE,
    )
    c.draw.rounded_rectangle(
        (980 * SCALE, 120 * SCALE, 1460 * SCALE, 900 * SCALE),
        radius=14 * SCALE,
        fill=(240, 238, 234),
        outline=LINE,
    )
    c.text((280, 146), "FRONTEND — React + Vite", f=font(BOLD, 14), fill=MUTED)
    c.text((1220, 146), "BACKEND — Go (biblioteca padrão)", f=font(BOLD, 14), fill=MUTED)

    front = [
        ("Componentes", ["App · Board · Column · TaskCard · TaskForm"], []),
        ("useTasks (hook)", ["Estado do quadro, loading/erro,", "atualizações otimistas + rollback"], []),
        ("api/client.js", ["fetch + tratamento de erro (ApiError)"], ["VITE_API_URL"]),
    ]
    y = 190
    for title, lines, mono in front:
        c.box((70, y, 490, y + 96), title=title, lines=lines, mono_lines=mono, accent=BLUE)
        if title != "api/client.js":
            c.arrow((280, y + 96), (280, y + 130))
        y += 130

    back = [
        ("Middlewares", ["recover → log de acesso → CORS"], ["Access-Control-Allow-*"]),
        ("http.ServeMux (Go 1.22)", ["Rotas por método e path"], ["GET|POST /api/tasks", "PUT|DELETE /api/tasks/{id}"]),
        ("handlers.go", ["Decodifica o JSON, valida e", "traduz erros em HTTP"], ["400 · 404 · 500"]),
        ("store.go — memória", ["Slice ordenado + RWMutex", "A ordem do slice é a ordem do quadro"], []),
    ]
    y = 190
    for title, lines, mono in back:
        height = 96 if not mono else 104
        c.box((1010, y, 1430, y + height), title=title, lines=lines, mono_lines=mono, accent=BRAND)
        c.arrow((1220, y + height), (1220, y + 130 - 4))
        y += 130

    c.box(
        (1010, y, 1430, y + 96),
        title="data/tasks.json (opcional)",
        lines=["Gravação atômica: arquivo temporário + rename", "DATA_FILE vazio = somente memória"],
        accent=GREEN,
    )

    # trocas HTTP
    c.text((750, 200), "HTTP / JSON", f=font(BOLD, 15), fill=MUTED)

    c.arrow((500, 250), (1000, 250), color=BLUE, label="requisição: método + corpo JSON")
    c.text(
        (750, 276),
        '{ "title": "…", "description": "…", "status": "todo" }',
        f=font(MONO, 12),
        fill=MUTED,
    )

    c.arrow((1000, 340), (500, 340), color=GREEN, label="resposta 2xx: tarefa ou lista de tarefas")
    c.text(
        (750, 366),
        '{ "id": "…", "status": "doing", "position": 0 }',
        f=font(MONO, 12),
        fill=MUTED,
    )

    c.arrow((1000, 430), (500, 430), color=RED, label="erro 4xx/5xx")
    c.text(
        (750, 456),
        '{ "error": { "code": "validation_error", "fields": { … } } }',
        f=font(MONO, 12),
        fill=MUTED,
    )

    c.box(
        (560, 520, 940, 626),
        title="CORS",
        lines=["O frontend (5173) e a API (8080) são", "origens distintas: o preflight OPTIONS", "é respondido com 204 pelo middleware."],
        accent=PURPLE,
        fill=WHITE,
    )

    c.box(
        (560, 660, 940, 790),
        title="Atualização otimista",
        lines=[
            "Mover, editar e excluir mudam a tela",
            "antes da resposta. Se a API falhar, o",
            "estado anterior é restaurado e a faixa",
            "de erro aparece no topo do quadro.",
        ],
        accent=BLUE,
        fill=WHITE,
    )

    c.save(path)


if __name__ == "__main__":
    here = os.path.dirname(os.path.abspath(__file__))
    user_flow(os.path.join(here, "user-flow.png"))
    data_flow(os.path.join(here, "data-flow.png"))
