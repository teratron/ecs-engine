# storage-sparse-set-go.md

Статья Рафиула Алама — это отличная демонстрация того, как появление дженериков (Generics) в Go 1.18+ позволило сделать ECS типобезопасным без использования медленных `interface{}` (boxing/unboxing) и рефлексии.

Однако, если мы строим `ecs-engine` с оглядкой на **Bevy** (AAA-уровень, максимальная производительность), подход из статьи нужно применять **с критическими оговорками**. Автор использует `map[Entity]T` для хранения компонентов и аллоцирует новые слайсы `make([]Entity, 0)` при каждом запросе (`Query`). Для высоконагруженного игрового цикла это «смертный приговор» из-за промахов кэша (map lookups) и пауз сборщика мусора (GC).

Вот готовый вариант **«Скилла / Набора правил» (AI Skill)**, который вы можете загрузить в системный промпт или добавить как `.agents/skill-go-generics-ecs.md`. Он заберет лучшие идеи из статьи, но заставит ИИ писать код, соответствующий строгим требованиям вашего движка.

---

## Файл для ИИ: `skill-ecs-generics-architecture.md`

**Описание:** Директивы для агента по реализации паттерна ECS с использованием Go Generics, основанные на концепции Type-Safety, но с жестким фокусом на Data-Oriented Design (DoD) и нулевых аллокациях в runtime.

### 🧠 Базовые принципы (Что брать из референсов)

1. **Типобезопасность (Type Safety):** Полностью исключить использование `any` (`interface{}`) для хранения и извлечения компонентов. Взаимодействие с данными должно идти через обобщенные структуры `ComponentStore[T any]`.
2. **Чистые данные:** Сущности (`Entity`) — это строго псевдонимы для `uint64` (или `uint32`). Компоненты — это чистые структуры (`struct`) без логики. Логика живет исключительно в Системах.
3. **Fluent API для Запросов:** Интерфейс построения запросов должен быть декларативным (наподобие `world.Query().With[Position]().With[Velocity]()`).

### 🚫 Критические ограничения (Как делать НЕЛЬЗЯ)

1. **ANTI-PATTERN: Maps для хранения.** Запрещено реализовывать `ComponentStore[T]` на базе `map[Entity]T`. Итерация по `map` в Go рандомизирована и убивает L1/L2 кэш-локальность.
2. **ANTI-PATTERN: Hardcoded World.** Запрещено хардкодить типы компонентов внутри структуры `World` (как `positions *ComponentStore[Position]`). `World` должен поддерживать динамическую регистрацию любых `T` компонентов при инициализации (через TypeID).
3. **ANTI-PATTERN: Runtime Allocations в Query.** Запрещено использовать `make([]Entity, 0)` или `append()` внутри методов `Query` или систем, которые выполняются каждый кадр (в цикле `Update`).

### 🛠️ Архитектурные Решения (Как ИИ ДОЛЖЕН писать код)

**Правило 1: Истинный Sparse Set (Хранение)**
Вместо мап ИИ обязан реализовать `ComponentStore[T]` в виде Sparse Set:

* `dense []T` — плотный массив самих данных (гарантирует линейное чтение из кэша процессора).
* `sparse []int` — разреженный массив, где индекс равен Entity ID, а значение — индексу в массиве `dense`.
* `entities []Entity` — плотный массив Entity ID, соответствующих компонентам в `dense` для обратного маппинга.

**Правило 2: Кэшированные запросы (Archetypes / Query Caching)**
Для реализации `Query().With[T]()` ИИ должен спроектировать механизм **предварительного кэширования**.

* Запросы не должны фильтровать сущности на лету каждый кадр.
* При добавлении/удалении компонента сущность должна автоматически перемещаться в соответствующие «архетипы» (Archetypes) или обновлять кэшированные списки (Views).
* Метод `Query().Execute()` должен возвращать уже готовый слайс или итератор по сплошному куску памяти, занимая $O(1)$ времени и $0$ аллокаций.

**Правило 3: Итераторы вместо возврата слайсов**
Для защиты памяти и минимизации копирования, системы должны получать доступ к данным через итераторы с передачей указателей на внутренние элементы `dense` массива:

```go
// Пример требуемого API (AI должен генерировать подобное):
query.ForEach(func(e Entity, pos *Position, vel *Velocity) {
    pos.X += vel.X * deltaTime
})
```

---

**Статус:** Утверждено для MVP
**Зависимости:** `entity-core.md`
**Цель:** Обеспечить кэш-локальное, типобезопасное хранилище компонентов O(1) без использования `map` и `interface{}`, минимизируя нагрузку на сборщик мусора (GC).

## 1. Архитектура данных (Memory Layout)

Хранилище реализует паттерн **Sparse Set** (Разреженное множество).
Структура состоит из трех синхронизированных массивов:

* `dense []T`: Плотный массив данных компонентов. Гарантирует линейное чтение из L1/L2 кэша. Никаких «дыр» в памяти.
* `entities []Entity`: Плотный массив идентификаторов сущностей. Индекс совпадает с индексом в `dense`. Необходим для обратного маппинга при удалении.
* `sparse []int32`: Разреженный массив. Индексом выступает сам `Entity ID`, а значением — индекс в плотных массивах (`dense` и `entities`).

## 2. Поведение и инварианты

* **Добавление (Add):** Компонент добавляется в конец `dense`, сущность в конец `entities`. В `sparse` по индексу `Entity` записывается длина `dense - 1`. Сложность: O(1) амортизированная.
* **Чтение (Get):** Прямое обращение: `dense[sparse[Entity]]`. Сложность: O(1).
* **Удаление (Remove - Swap & Pop):** 1. Находим индекс удаляемого элемента через `sparse`.
    2. Берем *последний* элемент из `dense` и `entities` и перемещаем его на место удаляемого.
    3. Обновляем индекс перемещенного элемента в `sparse`.
    4. Усекаем `dense` и `entities` на один элемент.
    Сложность: O(1). Порядок элементов не сохраняется (что является нормой для ECS).

## 3. Ограничения Go и Workarounds

* **Реаллокация слайсов:** При добавлении элементов встроенный `append` может выделить новый массив. Возвращать прямые указатели на элементы `dense` массива наружу хранилища (для долговременного хранения) **запрещено**, так как GC Go может сделать их невалидными или создать утечки.
* **Итерация:** Должна происходить строго через методы хранилища, либо через получение сырых массивов для локального использования в системах.

---

## 4. Эталонная реализация (MVP)

```go
package ecs

// Entity - базовый тип. В будущем можно разбить на ID (нижние 32 бита) и Generation (верхние 32 бита).
type Entity uint32

const InvalidIndex int32 = -1

// ComponentStore реализует кэш-дружелюбное хранилище для любого типа компонентов T.
type ComponentStore[T any] struct {
 dense    []T
 entities []Entity
 sparse   []int32
}

// NewComponentStore создает хранилище с заранее выделенной памятью для минимизации аллокаций.
func NewComponentStore[T any](initialCapacity int) *ComponentStore[T] {
 return &ComponentStore[T]{
  dense:    make([]T, 0, initialCapacity),
  entities: make([]Entity, 0, initialCapacity),
  sparse:   make([]int32, 0, initialCapacity),
 }
}

// Add добавляет или обновляет компонент для сущности.
func (s *ComponentStore[T]) Add(entity Entity, comp T) {
 idx := int(entity)
 
 // Динамическое расширение sparse массива без append в цикле
 if idx >= len(s.sparse) {
  newSparse := make([]int32, idx+1, (idx+1)*2)
  copy(newSparse, s.sparse)
  for i := len(s.sparse); i < len(newSparse); i++ {
   newSparse[i] = InvalidIndex
  }
  s.sparse = newSparse
 }

 denseIdx := s.sparse[idx]

 // Если компонент уже есть - просто обновляем данные
 if denseIdx != InvalidIndex {
  s.dense[denseIdx] = comp
  return
 }

 // Иначе добавляем в конец плотных массивов
 s.sparse[idx] = int32(len(s.dense))
 s.dense = append(s.dense, comp)
 s.entities = append(s.entities, entity)
}

// Get возвращает указатель на компонент для модификации на месте (in-place).
// ВАЖНО: Указатель валиден только до следующего Add или Remove!
func (s *ComponentStore[T]) Get(entity Entity) (*T, bool) {
 idx := int(entity)
 if idx >= len(s.sparse) {
  return nil, false
 }

 denseIdx := s.sparse[idx]
 if denseIdx == InvalidIndex {
  return nil, false
 }

 return &s.dense[denseIdx], true
}

// Remove удаляет компонент используя паттерн Swap and Pop.
func (s *ComponentStore[T]) Remove(entity Entity) {
 idx := int(entity)
 if idx >= len(s.sparse) {
  return
 }

 denseIdx := s.sparse[idx]
 if denseIdx == InvalidIndex {
  return
 }

 lastDenseIdx := len(s.dense) - 1

 // Если удаляем не последний элемент, меняем его местами с последним
 if int(denseIdx) != lastDenseIdx {
  lastEntity := s.entities[lastDenseIdx]
  
  // Перемещаем данные последнего элемента на место удаляемого
  s.dense[denseIdx] = s.dense[lastDenseIdx]
  s.entities[denseIdx] = lastEntity
  
  // Обновляем индекс перемещенного элемента в sparse массиве
  s.sparse[lastEntity] = denseIdx
 }

 // Инвалидируем индекс удаляемого элемента
 s.sparse[idx] = InvalidIndex

 // Усекаем слайсы (без аллокаций)
 var zero T // Обнуляем последний элемент для помощи GC (избегаем утечек памяти)
 s.dense[lastDenseIdx] = zero 
 
 s.dense = s.dense[:lastDenseIdx]
 s.entities = s.entities[:lastDenseIdx]
}

// Has проверяет наличие компонента.
func (s *ComponentStore[T]) Has(entity Entity) bool {
 idx := int(entity)
 return idx < len(s.sparse) && s.sparse[idx] != InvalidIndex
}

// Raw возвращает прямые слайсы для систем, которым нужна максимальная скорость итерации.
// Возвращаемые срезы имеют одинаковую длину.
func (s *ComponentStore[T]) Raw() ([]Entity, []T) {
 return s.entities, s.dense
}
