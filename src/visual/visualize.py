import math
import threading
import time
import grpc
import chord_message_pb2
import chord_message_pb2_grpc

import tkinter as tk
from tkinter import font
from concurrent import futures

MaxHashSlots = 1 << 6


class ChordListener(chord_message_pb2_grpc.ChordUpdateServicer):
    def __init__(self, vs):
        super().__init__()
        self.vs = vs

    def Update(self, request, context):
        if request.verb == 0:
            vs.update_node(request.nid)
            print("{} joins".format(request.nid))
        elif request.verb == 1:
            succ_id = int(request.value)
            vs.update_node(request.nid, successor=succ_id)
            print("{} set succ {}".format(request.nid, succ_id))
        elif request.verb == 2:
            pred_id = int(request.value)
            vs.update_node(request.nid, predecessor=pred_id)
            print("{} set pred {}".format(request.nid, pred_id))
        elif request.verb == 3:
            hi = request.value == "1"
            vs.update_node(request.nid, highlight=hi)
            print("{} set highlight {}".format(request.nid, hi))
        elif request.verb == 4:
            kv = {request.key: request.value}
            vs.update_node(request.nid, add_kv=kv)
            print("{} set key {}: \"{}\"".format(
                request.nid, request.key, request.value))
        elif request.verb == 5:
            vs.update_news(request.value)
            print("news: {}".format(request.value))


        return chord_message_pb2.ChordReply(reply=True)


class Chord:
    def __init__(self, id, predecessor=-1, successor=-1, highlight=False,
                 kv=None):
        self.id = id
        self.predecessor = predecessor
        self.successor = successor
        self.highlight = highlight
        self.kv = {}
        if kv is not None:
            self.kv = kv


class Ring:
    def __init__(self):
        self.ring = {}

    def update_node(self, nid, predecessor=None, successor=None, highlight=None,
                    add_kv=None):
        if nid in self.ring:
            if predecessor != None:
                self.ring[nid].predecessor = predecessor
            if successor is not None:
                self.ring[nid].successor = successor
            if highlight is not None:
                self.ring[nid].highlight = highlight
            if add_kv is not None:
                for k, v in add_kv.items():
                    self.ring[nid].kv[k] = v
        else:
            p, s = -1, -1
            h = False
            if predecessor is not None:
                p = predecessor
            if successor is not None:
                s = successor
            if highlight is not None:
                h = highlight
            self.ring[nid] = Chord(nid, predecessor=p, successor=s, highlight=h,
                                   kv=add_kv)

    def get_ring(self):
        ret = self.ring.values()
        sorted(ret, key=lambda ch: ch.id)
        return ret

    def get(self, nid):
        if nid in self.ring:
            return self.ring[nid]
        return None


class ChordDrawer:
    def __init__(self, canvas, width, height):
        self.canvas = canvas
        self.width = width
        self.height = height
        self.radius = 3 * min(width, height) / 8
        self.nid = None
        self.hl = False
        self.succ_id = None
        self.pred_id = None
        self.tk_node = None
        self.tk_text = None
        self.tk_succ = None
        self.tk_pred = None
        self.tk_hl = None
        self.tk_kv = {}

    def ring_loc(self, x, y, r, theta):
        x = math.floor(x + r * math.sin(theta))
        y = math.floor(y - r * math.cos(theta))
        return x, y

    def check(self, nid):
        return nid is not None and nid >= 0 and nid < MaxHashSlots

    def update(self, node):
        if not self.check(node.id):
            return

        theta = 2 * node.id * math.pi / MaxHashSlots
        x0, y0 = self.width // 2, self.height // 2
        cx, cy = self.ring_loc(x0, y0, self.radius, theta)
        cr = 10
        diva = 0.5

        # node plot and text plot
        if self.check(node.id) and self.nid != node.id:
            self.delete()
            self.nid = node.id
            self.tk_node = self.canvas.create_oval(cx - cr, cy - cr, cx + cr,
                                                   cy + cr, fill="yellow")

            tx, ty = self.ring_loc(cx, cy, 30, theta)
            self.tk_text = self.canvas.create_text(tx, ty,
                                                   text="{}".format(node.id),
                                                   font=font.Font(
                                                       family='Helvetica',
                                                       size=18,
                                                       weight='bold'))

        # highlight plot
        if self.tk_hl is not None and self.hl != node.highlight:
            self.canvas.delete(self.tk_hl)
            self.tk_hl = None
        if self.hl != node.highlight and node.highlight:
            self.tk_hl = self.canvas.create_oval(cx - cr, cy - cr, cx + cr,
                                                 cy + cr, fill="green")
        self.hl = node.highlight

        # successor pointer plot
        succ = node.successor
        if self.tk_succ is not None and self.succ_id != succ:
            self.canvas.delete(self.tk_succ)
            self.tk_succ = None

        if self.check(succ) and self.succ_id != succ:
            self.succ_id = succ
            sa = 2 * succ * math.pi / MaxHashSlots
            x, y = self.ring_loc(cx, cy, cr, theta + math.pi / 2 - diva)
            sx, sy = self.ring_loc(x0, y0, self.radius, sa)
            sx, sy = self.ring_loc(sx, sy, cr, sa - math.pi / 2 + diva)
            self.tk_succ = self.canvas.create_line(x, y, sx, sy, fill="red",
                                                   arrow=tk.LAST)

        # predecessor pointer plot
        pred = node.predecessor
        if self.tk_pred is not None and self.pred_id != pred:
            self.canvas.delete(self.tk_pred)
            self.tk_pred = None

        if self.check(pred) and self.pred_id != pred:
            self.pred_id = pred
            pa = 2 * pred * math.pi / MaxHashSlots
            x, y = self.ring_loc(cx, cy, cr, theta - math.pi / 2 - diva)
            px, py = self.ring_loc(x0, y0, self.radius, pa)
            px, py = self.ring_loc(px, py, cr, pa + math.pi / 2 + diva)
            self.tk_pred = self.canvas.create_line(x, y, px, py, fill="blue",
                                                   arrow=tk.LAST)

        # kv plot
        is_sub = all([k in node.kv for k in self.tk_kv.keys()])
        if not is_sub:
            for k in self.tk_kv.keys():
                self.canvas.delete(self.tk_kv[k])
            self.tk_kv = {}

        for k, v in node.kv.items():
            if k in self.tk_kv:
                self.canvas.itemconfigure(self.tk_kv[k],
                                          text="{}: \"{}\"".format(k, v))
            else:
                kx, ky = self.ring_loc(cx, cy, 4*cr, theta+math.pi)
                ky += (2*(len(self.tk_kv)%2)-1) * ((len(self.tk_kv)+1)//2) * 12
                plotk = self.canvas.create_text(
                    kx, ky,
                    text="{}: \"{}\"".format(k, v),
                    font=font.Font(
                        family='Helvetica',
                        size=12,
                        weight='bold',
                    ),
                    fill="#696969"
                )
                self.tk_kv[k] = plotk

    def delete(self):
        if self.tk_node is not None:
            self.canvas.delete(self.tk_node)
        if self.tk_text is not None:
            self.canvas.delete(self.tk_text)
        if self.tk_pred is not None:
            self.canvas.delete(self.tk_pred)
        if self.tk_succ is not None:
            self.canvas.delete(self.tk_succ)
        if self.tk_hl is not None:
            self.canvas.delete(self.tk_hl)
        for k in self.tk_kv.keys():
            if self.tk_kv[k] is not None:
                self.canvas.delete(self.tk_kv[k])


class Visualizer(tk.Tk):
    def __init__(self, width, height, listen_addr="localhost:6008"):
        super().__init__()
        self.addr = listen_addr
        self.ring = Ring()
        self.width = width
        self.height = height
        self.canvas = tk.Canvas(self, width=800, height=600, bg='white')
        self.canvas.pack()
        self.draws = {}
        self.news = "Latest event:"
        self.tk_news = None
        self.title("You see the Chord!")
        self.bind('<Destroy>', self.kill)

        # plot ring
        w, h = self.width, self.height
        x0, y0 = w / 2, h / 2
        r = 3 * min(w, h) / 8
        self.canvas.create_oval(x0 - r, y0 - r, x0 + r, y0 + r, dash=(5, 5))

        self.worker = threading.Thread(target=self.listen, name="update")
        self.cond = threading.Condition()
        self.stop = False

        self.worker.start()

    def update_news(self, news):
        with self.cond:
            if news == self.news:
                return
            self.news = news
            if self.tk_news is not None:
                self.canvas.delete(self.tk_news)
            self.tk_news = self.canvas.create_text(
                self.width*8//10,
                self.height*9//10,
                text="Latest event: {}".format(self.news),
                font=font.Font(family='Helvetica', size=18, weight='bold')
            )
            self.canvas.update()

    def update_node(self, nid, predecessor=None, successor=None, highlight=None,
                    add_kv=None):
        with self.cond:
            self.ring.update_node(nid, predecessor=predecessor,
                                  successor=successor, highlight=highlight,
                                  add_kv=add_kv)
            if nid not in self.draws:
                self.draws[nid] = ChordDrawer(self.canvas, self.width,
                                              self.height)
            self.draws[nid].update(self.ring.ring[nid])
            self.canvas.update()

    def listen(self):
        server = grpc.server(futures.ThreadPoolExecutor(max_workers=1))
        chord_message_pb2_grpc.add_ChordUpdateServicer_to_server(
            ChordListener(self), server)

        print("Listen Chord news on: {}".format(self.addr))
        server.add_insecure_port(self.addr)
        server.start()

        try:
            while True:
                with self.cond:
                    if self.stop:
                        server.stop(0)
                        print("Stop listening {}".format(self.addr))
                        return
                    self.cond.wait()
        except KeyboardInterrupt:
            server.stop(0)

    def kill(self, event):
        with self.cond:
            self.stop = True
            self.cond.notify()


if __name__ == "__main__":
    vs = Visualizer(800, 600)
    vs.mainloop()
    vs.worker.join()
