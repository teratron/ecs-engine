# Go Game Engine — Conventions & Rules

> Стиль, архитектура и паттерны для чистого, идиоматичного Go-движка.

---

## 1. Структура пакетов

```
engine/
├── cmd/
│   └── game/
│       └── main.go          # точка входа, только инициализация
├── internal/
│   ├── core/                # ECS: World, Entity, Component, System
│   ├── renderer/            # рендеринг (OpenGL / Vulkan / WebGPU)
│   ├── physics/             # физика и коллизии
│   ├── audio/               # звук
│   ├── input/               # обработка ввода
│   ├── asset/               # загрузка ресурсов
│   └── math/                # векторы, матрицы, кватернионы
├── pkg/
│   └── ecs/                 # публичный API ECS (если библиотека)
├── game/
│   ├── scenes/              # сцены / уровни
│   ├── components/          # игровые компоненты
│   └── systems/             # игровые системы
└── assets/
    ├── shaders/
    ├── textures/
    └── sounds/
```

**Правила:**
- `internal/` — движок; `game/` — игровая логика. Никогда не смешивать.
- Один пакет = одна ответственность. Пакет `renderer` не знает про `physics`.
- `cmd/game/main.go` содержит только `main()` — никакой логики.

---

## 2. Именование

### Пакеты
```go
// ПРАВИЛЬНО: короткое, в нижнем регистре, без подчёркиваний
package ecs
package renderer
package math

// НЕПРАВИЛЬНО
package gameEngine
package game_engine
package GameEngine
```

### Типы и интерфейсы
```go
// Интерфейс: описывает поведение, часто суффикс -er
type Renderer interface { ... }
type Updater  interface { Update(dt float64) }
type System   interface { Update(world *World, dt float64) }

// Конкретные типы: существительное
type MeshRenderer struct { ... }
type RigidBody    struct { ... }
type Transform    struct { ... }
```

### Конструкторы
```go
// New{Type} — всегда, без исключений
func NewWorld() *World            { ... }
func NewEntity(id EntityID) Entity { ... }
func NewTransform() *Transform    { ... }
```

### Константы и перечисления
```go
// Группировать через iota, с префиксом типа
type RenderMode int

const (
    RenderModeForward RenderMode = iota
    RenderModeDeferred
    RenderModeRayTrace
)
```

---

## 3. ECS — Entity Component System

### EntityID — value type, не указатель
```go
// ПРАВИЛЬНО
type EntityID uint64

const InvalidEntity EntityID = 0

// Entity — просто ID, без данных
type Entity struct {
    ID         EntityID
    Generation uint32
}
```

### Component — чистые данные, без методов с логикой
```go
// ПРАВИЛЬНО: только данные
type Transform struct {
    Position Vec3
    Rotation Quat
    Scale    Vec3
}

// НЕПРАВИЛЬНО: логика в компоненте
func (t *Transform) UpdateMatrix() { ... } // → это задача системы
```

### System — вся логика, без состояния игры
```go
type System interface {
    Update(world *World, dt float64)
}

type PhysicsSystem struct {
    gravity Vec3
    // внутреннее состояние системы — ок
    broadphase *BroadPhase
}

func (s *PhysicsSystem) Update(world *World, dt float64) {
    // итерируем по сущностям через Query
    world.Query(
        With[RigidBody](),
        With[Transform](),
    ).Each(func(e Entity, rb *RigidBody, tr *Transform) {
        // физика
    })
}
```

### World — центральный реестр
```go
type World struct {
    entities   EntityManager
    components ComponentRegistry
    systems    []System
    // запрещено хранить состояние рендера, физики и т.д.
}

func (w *World) AddSystem(s System)              { w.systems = append(w.systems, s) }
func (w *World) Update(dt float64)               { for _, s := range w.systems { s.Update(w, dt) } }
func (w *World) Spawn() Entity                   { return w.entities.Create() }
func (w *World) Despawn(e Entity)                { w.entities.Destroy(e) }
func (w *World) Add(e Entity, c ...Component)    { ... }
func (w *World) Get[T Component](e Entity) (*T, bool) { ... }
```

---

## 4. Game Loop

```go
// internal/core/loop.go

type GameLoop struct {
    targetTPS   int           // ticks per second (логика)
    targetFPS   int           // frames per second (рендер)
    world       *World
    renderer    Renderer
    running     atomic.Bool
}

func (l *GameLoop) Run(ctx context.Context) error {
    tickInterval  := time.Second / time.Duration(l.targetTPS)
    frameInterval := time.Second / time.Duration(l.targetFPS)

    var (
        lastTick  = time.Now()
        lastFrame = time.Now()
    )

    for l.running.Load() {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
        }

        now := time.Now()

        if now.Sub(lastTick) >= tickInterval {
            dt := now.Sub(lastTick).Seconds()
            l.world.Update(dt)
            lastTick = now
        }

        if now.Sub(lastFrame) >= frameInterval {
            l.renderer.Frame(l.world)
            lastFrame = now
        }
    }
    return nil
}
```

**Правила:**
- Логика и рендер — независимые частоты обновления.
- `ctx context.Context` — обязательный аргумент для корректного завершения.
- Никаких `time.Sleep` в основном цикле.

---

## 5. Обработка ошибок

```go
// ПРАВИЛЬНО: оборачивать с контекстом
func (l *AssetLoader) LoadTexture(path string) (*Texture, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("asset.LoadTexture %q: %w", path, err)
    }
    // ...
}

// ПРАВИЛЬНО: sentinel errors для проверки на стороне вызывающего
var (
    ErrEntityNotFound   = errors.New("entity not found")
    ErrComponentMissing = errors.New("component not attached")
    ErrInvalidAsset     = errors.New("invalid asset format")
)

// ПРАВИЛЬНО: panic только для невосстановимых ошибок программирования
func MustLoadShader(src string) *Shader {
    s, err := CompileShader(src)
    if err != nil {
        panic(fmt.Sprintf("shader compilation failed: %v", err))
    }
    return s
}
```

**Правила:**
- Никогда не возвращать `error` без контекста (`fmt.Errorf(...%w...)`)
- `panic` — только для ошибок программирования, не для ошибок рантайма
- `Must*` функции используются только при инициализации, не в игровом цикле

---

## 6. Производительность и память

### Избегать аллокаций в горячем пути
```go
// НЕПРАВИЛЬНО: аллокация каждый кадр
func (s *RenderSystem) Update(world *World, dt float64) {
    transforms := make([]Transform, 0)  // ← аллокация!
    // ...
}

// ПРАВИЛЬНО: переиспользовать срез
type RenderSystem struct {
    transformCache []Transform // переиспользуется каждый кадр
}

func (s *RenderSystem) Update(world *World, dt float64) {
    s.transformCache = s.transformCache[:0] // сброс без аллокации
    // ...
}
```

### Object Pool для часто создаваемых объектов
```go
var particlePool = sync.Pool{
    New: func() any { return &Particle{} },
}

func SpawnParticle() *Particle {
    p := particlePool.Get().(*Particle)
    p.Reset()
    return p
}

func DespawnParticle(p *Particle) {
    particlePool.Put(p)
}
```

### SoA (Structure of Arrays) для компонентов с интенсивной итерацией
```go
// НЕПРАВИЛЬНО для большого количества сущностей: AoS
type Entity struct {
    Position Vec3
    Velocity Vec3
    // ...
}

// ПРАВИЛЬНО: SoA — cache-friendly
type PhysicsStorage struct {
    positions  []Vec3
    velocities []Vec3
    masses     []float32
}
```

---

## 7. Конкурентность

```go
// Системы обновляются последовательно по умолчанию
// Параллелизм — только для независимых систем через errgroup

func (w *World) UpdateParallel(dt float64) error {
    g, ctx := errgroup.WithContext(context.Background())
    _ = ctx

    // только системы, помечённые как независимые
    for _, s := range w.parallelSystems {
        s := s
        g.Go(func() error {
            s.Update(w, dt)
            return nil
        })
    }
    return g.Wait()
}
```

**Правила:**
- Никакого shared state между системами без явной синхронизации.
- `sync.Mutex` — для редких операций. `sync/atomic` — для счётчиков.
- Каналы — для передачи событий между горутинами, не для передачи данных компонентов.
- Горутины всегда завершаются через `context.Context` или close-channel.

---

## 8. Интерфейсы и зависимости

```go
// ПРАВИЛЬНО: интерфейс определяется на стороне потребителя
// в пакете renderer — только то, что нужно рендереру
package renderer

type TextureSource interface {
    Pixels() []byte
    Width() int
    Height() int
}

// НЕПРАВИЛЬНО: интерфейс в пакете производителя
package asset
type TextureInterface interface { ... } // ← не делать так
```

**Принцип:** интерфейсы маленькие (1–3 метода). Большие интерфейсы — признак плохого дизайна.

---

## 9. Математика (векторы, матрицы)

```go
// Значимые типы (value types), не указатели
type Vec2 struct{ X, Y float32 }
type Vec3 struct{ X, Y, Z float32 }
type Vec4 struct{ X, Y, Z, W float32 }
type Mat4 [16]float32
type Quat struct{ X, Y, Z, W float32 }

// Методы возвращают новое значение — immutable
func (v Vec3) Add(u Vec3) Vec3    { return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z} }
func (v Vec3) Scale(s float32) Vec3 { return Vec3{v.X * s, v.Y * s, v.Z * s} }
func (v Vec3) Dot(u Vec3) float32  { return v.X*u.X + v.Y*u.Y + v.Z*u.Z }
func (v Vec3) Normalize() Vec3 {
    inv := 1.0 / float32(math.Sqrt(float64(v.Dot(v))))
    return v.Scale(inv)
}
```

---

## 10. Логирование и отладка

```go
// Использовать стандартный log/slog (Go 1.21+)
import "log/slog"

var logger = slog.Default()

func (s *PhysicsSystem) Update(world *World, dt float64) {
    slog.Debug("physics update", "dt", dt, "entities", world.EntityCount())
    // ...
}

// Уровни: Debug — разработка, Info — события, Warn — проблемы, Error — ошибки
```

---

## 11. Тестирование

```go
// Тесты рядом с кодом: math/vec3_test.go
func TestVec3Normalize(t *testing.T) {
    v := Vec3{3, 0, 0}
    got := v.Normalize()
    want := Vec3{1, 0, 0}
    if got != want {
        t.Errorf("Normalize() = %v, want %v", got, want)
    }
}

// Benchmark для горячих путей
func BenchmarkVec3Dot(b *testing.B) {
    v, u := Vec3{1, 2, 3}, Vec3{4, 5, 6}
    for b.Loop() {
        v.Dot(u)
    }
}

// Для систем — mock через интерфейс
type mockRenderer struct{ frameCount int }
func (m *mockRenderer) Frame(*World)    { m.frameCount++ }
func (m *mockRenderer) Shutdown() error { return nil }
```

---

## 12. Чеклист перед коммитом

- [ ] `go vet ./...` — без предупреждений
- [ ] `staticcheck ./...` — без предупреждений  
- [ ] `go test ./...` — все тесты зелёные
- [ ] Нет аллокаций в функциях горячего пути (`go test -bench -benchmem`)
- [ ] Новые публичные типы/функции — задокументированы (`// TypeName ...`)
- [ ] Интерфейсы определены у потребителя, не у производителя
- [ ] Ошибки обёрнуты через `fmt.Errorf("контекст: %w", err)`
- [ ] Горутины завершаются корректно (context/channel)

---

## Инструменты

| Инструмент | Назначение |
|---|---|
| `gopls` | LSP сервер |
| `staticcheck` | расширенный линтер |
| `golangci-lint` | мета-линтер |
| `go tool pprof` | профилирование CPU/памяти |
| `go test -race` | детектор гонок |
| `delve` | отладчик |
