server aloha {
    state {}

    receive fn hello() {
        print(:hellord)
    }

    receive fn bo_dia(grupo: string) {
        print(grupo)
    }

    receive fn add(n: number, a: number) {
        return n + a
    }

    receive fn worker(lb) {
        server worker {
            receive work() {
                return lb()
            }
        }
    }
}
