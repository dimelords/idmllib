# CLI TUI Architecture

## Overview

The idmlbuild CLI has been completely rewritten using [Bubbletea](https://github.com/charmbracelet/bubbletea), a modern TUI (Terminal User Interface) framework. This document explains the architecture, design decisions, and patterns used.

## Why Bubbletea?

**Previous Implementation:** Text-based with `bufio.Reader` for input  
**New Implementation:** Full TUI with bubbletea

### Benefits

1. **Modern UX** - Visual navigation, highlighting, and feedback
2. **Consistency** - Unified interaction patterns across all features
3. **Extensibility** - Easy to add new wizards and components
4. **Maintainability** - Clean separation of concerns
5. **Testability** - Components can be tested independently
6. **Professional** - Polished appearance with lipgloss styling

## Architecture Pattern

### The Elm Architecture

Bubbletea follows The Elm Architecture (TEA), a functional pattern with three core concepts:

1. **Model** - Application state
2. **Update** - State transformation based on messages
3. **View** - Render current state to string

```go
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

### Component Hierarchy

```
Main Loop
└── MainMenu (root selector)
    ├── CreateDocumentWizard
    │   ├── Step 1: Preset Selection
    │   ├── Step 2: Orientation Selection
    │   ├── Step 3: Columns Input
    │   ├── Step 4: Gutter Input
    │   ├── Step 5: Filename Input
    │   └── Step 6: Creation & Results
    │
    ├── RoundtripWizard
    │   ├── Step 1: Input File
    │   ├── Step 2: Output File
    │   └── Step 3: Processing & Results
    │
    └── ExportIDMSWizard
        ├── Step 1: Input File
        ├── Step 2: TextFrameSelector (sub-component)
        ├── Step 3: Output File
        └── Step 4: Export & Results
```

## Core Components

### 1. Main Menu (`main_menu.go`)

**Purpose:** Root-level menu selector  
**Pattern:** List selection  

**State:**
- `items []MenuItem` - Menu options
- `cursor int` - Currently highlighted item
- `selected int` - User's selection

**Key Features:**
- Quick number selection (1-3)
- Vim bindings (k/j)
- Arrow key navigation
- Dynamic descriptions

### 2. Create Document Wizard (`create_document.go`)

**Purpose:** Multi-step document creation  
**Pattern:** State machine wizard  

**States (Steps):**
```go
const (
    stepPreset      // Page size selection
    stepOrientation // Portrait/Landscape
    stepColumns     // Column count
    stepGutter      // Column spacing
    stepFilename    // Output path
    stepCreating    // Processing
    stepDone        // Results
)
```

**State:**
- `step` - Current wizard step
- `preset, orientation, columns, gutter, filename` - User choices
- `error` - Error message (if any)
- `success` - Success flag
- `pkg *idml.Package` - Created document

**Navigation:**
- `Esc` - Go back one step
- `Enter` - Advance to next step
- `Q` - Cancel wizard

### 3. Roundtrip Wizard (`roundtrip.go`)

**Purpose:** File integrity testing  
**Pattern:** Linear workflow  

**States:**
```go
const (
    rtStepInputFile   // Get input path
    rtStepOutputFile  // Get output path
    rtStepProcessing  // Execute roundtrip
    rtStepDone        // Show results
)
```

**Features:**
- Inline text input with cursor
- Default value support
- Progress indication
- Detailed statistics display

### 4. Export IDMS Wizard (`export_idms.go`)

**Purpose:** IDMS snippet export  
**Pattern:** Composite wizard (includes sub-component)  

**States:**
```go
const (
    expStepInputFile    // Get IDML path
    expStepSelectFrame  // TextFrameSelector
    expStepOutputFile   // Get IDMS path
    expStepExporting    // Process export
    expStepDone         // Show results
)
```

**Delegation Pattern:**
When in `expStepSelectFrame`, updates are delegated to the `TextFrameSelector` sub-component:

```go
if m.step == expStepSelectFrame {
    updatedSelector, cmd := m.frameSelector.Update(msg)
    m.frameSelector = updatedSelector.(*TextFrameSelector)
    return m, cmd
}
```

### 5. TextFrame Selector (`textframe_selector.go`)

**Purpose:** Visual textframe browser  
**Pattern:** List browser with preview  

**State:**
- `items []TextFrameItem` - All textframes
- `cursor int` - Highlighted frame
- `selected int` - User's selection
- `pkg, stories, spreads` - IDML data

**Features:**
- Story content preview (100 char)
- Frame and story ID display
- Keyboard navigation
- Alt-screen mode support

## Shared Components

### Styles (`styles.go`)

Centralized lipgloss styles for consistency:

```go
var (
    TitleStyle      // Bold, purple, padded
    SubtitleStyle   // Gray, subtle
    SelectedStyle   // Bold, highlighted background
    NormalStyle     // Default text
    SuccessStyle    // Green, bold
    ErrorStyle      // Red, bold
    WarningStyle    // Orange, bold
    InfoStyle       // Blue
    HelpStyle       // Gray, bottom padding
    InputStyle      // Input field appearance
    InputLabelStyle // Input label
    BoxStyle        // Bordered container
)
```

### Text Input (`text_input.go`)

**Purpose:** Reusable text input component  
**State:** prompt, placeholder, value, cursor  
**Features:** Real-time editing, backspace support  

**Usage:**
```go
input := tui.NewTextInput("Enter filename:", "output.idml")
p := tea.NewProgram(input)
m, _ := p.Run()
finalInput := m.(*tui.TextInput)
value := finalInput.GetValue()
```

## Design Patterns

### 1. Wizard Pattern

Multi-step workflows use enum-based state machines:

```go
type Step int

const (
    step1 Step = iota
    step2
    step3
    stepDone
)

type Wizard struct {
    step Step
    // ... fields for each step
}

func (w *Wizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        if msg.String() == "enter" {
            switch w.step {
            case step1:
                w.step = step2
            case step2:
                w.step = step3
            // ...
            }
        }
    }
    return w, nil
}
```

### 2. Component Delegation

Complex wizards can embed and delegate to sub-components:

```go
type ParentWizard struct {
    subComponent *ChildComponent
}

func (p *ParentWizard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if shouldDelegateToChild {
        updated, cmd := p.subComponent.Update(msg)
        p.subComponent = updated.(*ChildComponent)
        return p, cmd
    }
    // handle own updates
}
```

### 3. Result Display Pattern

Success screens use boxed summaries:

```go
func (w *Wizard) viewSuccess() string {
    var s strings.Builder
    s.WriteString(SuccessStyle.Render("✅ Success!"))
    s.WriteString("\n\n")
    
    details := fmt.Sprintf(
        "Info:\n"+
        "   Field 1: %s\n"+
        "   Field 2: %s\n",
        w.field1, w.field2,
    )
    
    s.WriteString(BoxStyle.Render(details))
    return s.String()
}
```

### 4. Progressive Enhancement

Graceful degradation for terminal limitations:

```go
// Use alt-screen for immersive experiences
p := tea.NewProgram(model, tea.WithAltScreen())

// Regular screen for simple flows
p := tea.NewProgram(model)
```

## Data Flow

### 1. Menu Selection Flow

```
User runs CLI
    ↓
MainMenu displays
    ↓
User selects option (1-3)
    ↓
Menu returns selected index
    ↓
Main loop launches appropriate wizard
    ↓
Wizard completes
    ↓
Returns to MainMenu
```

### 2. Wizard Flow

```
Wizard initializes (Init)
    ↓
User interacts (Update receives KeyMsg)
    ↓
State transitions (step changes)
    ↓
View renders current step
    ↓
... repeat until done
    ↓
Final state reached
    ↓
Wizard quits (tea.Quit)
    ↓
Control returns to main loop
```

### 3. Message Types

```go
// Input from user
type KeyMsg

// Terminal size change
type WindowSizeMsg

// Custom messages
type CustomMsg struct{ data string }

// Commands generate messages
func fetchData() tea.Msg {
    // ... fetch data
    return CustomMsg{data: result}
}
```

## State Management

### Immutability

Bubbletea encourages immutability - return new state rather than mutating:

```go
// ❌ Bad - mutates state
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.value = "new"  // mutation
    return m, nil
}

// ✅ Good - returns new state
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    newModel := *m
    newModel.value = "new"
    return &newModel, nil
}

// ✅ Also good - pointer receiver, explicit return
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    m.value = "new"
    return m, nil
}
```

### Pointer vs Value Receivers

**We use pointer receivers** (`*Model`) for Update and View because:
1. More efficient (no copying)
2. Clearer mutation semantics
3. Consistent with Init

## Error Handling

### Pattern

Errors are stored in wizard state and displayed in View:

```go
type Wizard struct {
    error   string
    success bool
}

func (w *Wizard) performAction() (tea.Model, tea.Cmd) {
    if err := doSomething(); err != nil {
        w.error = err.Error()
        w.step = stepDone
        return w, nil
    }
    w.success = true
    w.step = stepDone
    return w, nil
}

func (w *Wizard) View() string {
    if w.error != "" {
        return ErrorStyle.Render("❌ " + w.error)
    }
    if w.success {
        return w.viewSuccess()
    }
    // ... normal view
}
```

## Testing Strategy

### Unit Testing Models

```go
func TestWizardNavigation(t *testing.T) {
    w := NewWizard()
    
    // Test initial state
    if w.step != step1 {
        t.Error("Wrong initial step")
    }
    
    // Simulate key press
    msg := tea.KeyMsg{Type: tea.KeyEnter}
    updated, _ := w.Update(msg)
    
    finalWizard := updated.(*Wizard)
    if finalWizard.step != step2 {
        t.Error("Should advance to step2")
    }
}
```

### Integration Testing

```go
func TestFullFlow(t *testing.T) {
    w := NewWizard()
    
    // Simulate full interaction sequence
    msgs := []tea.Msg{
        tea.KeyMsg{Type: tea.KeyDown},   // Navigate
        tea.KeyMsg{Type: tea.KeyEnter},  // Select
        // ... more messages
    }
    
    for _, msg := range msgs {
        w, _ = w.Update(msg).(* Wizard)
    }
    
    // Verify final state
    if !w.success {
        t.Error("Flow should succeed")
    }
}
```

## Performance Considerations

### Efficient Rendering

- View only renders visible content
- No computation in View (preparation in Update)
- Reuse strings.Builder
- Cache expensive computations

### Memory Management

- Models are small (mostly primitives)
- Large data (IDML packages) passed by reference
- Components cleaned up when done

## Future Enhancements

### Potential Features

1. **Search/Filter** - Add filtering to TextFrame selector
2. **Multi-Select** - Select multiple frames for batch export
3. **Progress Bars** - For long operations
4. **Spinners** - For async tasks
5. **Forms** - Complex multi-field input
6. **Tables** - Tabular data display
7. **Tree Views** - Hierarchical data browsing

### Component Library

Build reusable components:
- `FileInput` - File path input with validation
- `NumberInput` - Numeric input with bounds
- `Selector` - Generic list selector
- `Confirmation` - Yes/No prompts
- `ProgressBar` - Progress indicator

## Best Practices

### 1. Single Responsibility

Each component should have one clear purpose.

### 2. Composition Over Inheritance

Embed sub-components rather than extending models.

### 3. Clear State Transitions

Use enums for steps, document state machine.

### 4. Consistent Styling

Always use shared styles from `styles.go`.

### 5. Keyboard Shortcuts

Support both arrow keys and vim bindings.

### 6. Help Text

Always show available keyboard shortcuts.

### 7. Error Recovery

Provide clear error messages and recovery paths.

### 8. Accessibility

Use clear labels and high-contrast colors.

## Resources

- [Bubbletea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Lipgloss Docs](https://github.com/charmbracelet/lipgloss)
- [Charm Examples](https://github.com/charmbracelet/bubbletea/tree/master/examples)
- [The Elm Architecture](https://guide.elm-lang.org/architecture/)

## Migration Guide

### From Old CLI to New TUI

**Old Pattern:**
```go
fmt.Print("Enter value: ")
value, _ := reader.ReadString('\n')
value = strings.TrimSpace(value)
```

**New Pattern:**
```go
input := tui.NewTextInput("Enter value:", "default")
p := tea.NewProgram(input)
m, _ := p.Run()
value := m.(*tui.TextInput).GetValue()
```

**Old Pattern:**
```go
fmt.Println("1. Option A")
fmt.Println("2. Option B")
fmt.Print("Choice: ")
choice, _ := reader.ReadString('\n')
```

**New Pattern:**
```go
menu := tui.NewMenu([]MenuItem{
    {Title: "Option A"},
    {Title: "Option B"},
})
p := tea.NewProgram(menu)
m, _ := p.Run()
selected := m.(*tui.Menu).GetSelected()
```

## Conclusion

The TUI rewrite provides a modern, maintainable, and extensible foundation for the idmlbuild CLI. The bubbletea architecture makes it easy to add new features while maintaining consistency and user experience quality.
