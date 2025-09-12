# HTMX & Hyperscript Patterns

A collection of proven patterns for building dynamic web applications with HTMX and Hyperscript, learned from building production applications with the DevTools framework.

## Core Principles

### HATEOAS (Hypermedia as the Engine of Application State)
- **Server-driven UI**: All state and logic lives on the server
- **HTML as state transfer**: Server returns HTML, not JSON
- **Progressive enhancement**: Works without JavaScript
- **No client-side routing**: All navigation is server-side

## HTMX Patterns

### Form Handling
```html
<!-- Use hx-target only for error containers, not the whole page -->
<div class="error-message"></div>
<form hx-post="/path" 
      hx-target="previous .error-message" 
      hx-swap="innerHTML">
    <!-- form fields -->
</form>
```

### Auto-refresh for Real-time Updates
```html
<!-- Stats that refresh every 10 seconds -->
<div id="stats-container" 
     hx-get="/partials/stats" 
     hx-trigger="every 10s" 
     hx-swap="innerHTML">
    <!-- Initial content with skeleton loaders -->
</div>
```

### Lazy Loading
```html
<!-- Load heavy content only when visible -->
<div hx-get="/heavy-content" 
     hx-trigger="revealed" 
     hx-swap="outerHTML">
    <!-- Skeleton loader -->
    <div class="animate-pulse">Loading...</div>
</div>
```

### Modal Dialogs
```html
<!-- Button to open modal -->
<button onclick="my_modal.showModal()">Open Modal</button>

<!-- Modal with HTMX form -->
<dialog id="my_modal" class="modal">
    <div class="modal-box">
        <form hx-post="/action" 
              hx-trigger="submit"
              _="on htmx:afterRequest call my_modal.close()">
            <!-- form content -->
        </form>
    </div>
    <!-- Click backdrop to close -->
    <form method="dialog" class="modal-backdrop">
        <button>close</button>
    </form>
</dialog>
```

## Hyperscript Patterns

### Basic Event Handling
```html
<!-- Simple click handler -->
<button _="on click toggle .hidden on #target">Toggle</button>

<!-- Conditional logic -->
<button _="on click 
           if #checkbox.checked 
               add .active to me
           else
               remove .active from me
           end">
    Conditional Button
</button>
```

### Dynamic Text Updates
```html
<!-- Change button text based on state -->
<button id="nav-button"
        _="on click
           set step to @data-step of #container
           if step == 1
               set me.innerText to 'Close'
           else
               set me.innerText to 'Previous'
           end">
    Previous
</button>
```

### Working with Data Attributes
```html
<!-- Store and retrieve data -->
<div id="container" data-step="1"
     _="on load
        set @data-step of me to 1
        send updateUI to me
     end">
</div>
```

### Multi-step Forms/Tours
```html
<script type="text/hyperscript">
    on showStep(step) from #tour-content
        set @data-step of #tour-content to step
        
        -- Hide all steps and show current
        add .hidden to .step-content
        set stepId to 'step-' + step
        remove .hidden from #{stepId}
        
        -- Update progress indicators
        for i in [1, 2, 3, 4]
            set indicatorId to 'indicator-' + i
            if step >= i
                add .active to #{indicatorId}
            else
                remove .active from #{indicatorId}
            end
        end
    end
</script>
```

## Common Gotchas & Solutions

### 1. HTML Entity Escaping
**Problem**: `<=` becomes `&lt;=` in templates
**Solution**: Reverse comparisons: `if step >= i` instead of `if i <= step`

### 2. Template Literals
**Problem**: Hyperscript doesn't support `{variable}` syntax
**Solution**: Use string concatenation: `'prefix-' + variable`

### 3. Comments in Hyperscript
**Problem**: `--` comments can cause parse errors
**Solution**: Remove all comments from production Hyperscript

### 4. Controller Methods in Go
```go
// Always use c.Redirect for HTMX compatibility
c.Redirect(w, r, "/path")       // ✅ Correct - HTMX-aware
http.Redirect(w, r, "/path", 302) // ❌ Wrong - breaks HTMX

// Refresh for HTMX partial updates
c.Refresh(w, r)  // Triggers page refresh via HX-Refresh header
```

### 5. Preventing Flicker
```html
<!-- Add min-height to containers that will be replaced -->
<div class="min-h-[320px]" hx-get="/content" hx-trigger="load">
    <!-- Skeleton loaders prevent layout shift -->
    <div class="animate-pulse">
        <div class="h-4 bg-base-300 rounded w-3/4 mb-2"></div>
        <div class="h-3 bg-base-300 rounded w-1/2"></div>
    </div>
</div>
```

### 6. Form Values in HTMX
```html
<!-- Include specific form fields in requests -->
<button hx-post="/save"
        hx-include="#field1,#field2">
    Save Selected
</button>

<!-- Or include by name attribute -->
<button hx-post="/save?dont_show=true"
        hx-include="[name='preferences']">
    Save with Preferences
</button>
```

## Best Practices

1. **Keep Hyperscript Simple**: Complex logic belongs on the server
2. **Use Semantic HTML**: Leverage proper form elements and buttons
3. **Progressive Enhancement**: Ensure basic functionality without JS
4. **Minimize Round Trips**: Batch related updates when possible
5. **Provide Visual Feedback**: Use loading states and transitions
6. **Handle Errors Gracefully**: Always have error containers ready

## Integration with DevTools

The DevTools `Controller` provides built-in HTMX support:

```go
// In your controller
func (c *MyController) Handler(w http.ResponseWriter, r *http.Request) {
    // Process form...
    
    // For HTMX requests, trigger a refresh
    c.Refresh(w, r)
    
    // Or render a partial
    if r.Header.Get("HX-Request") == "true" {
        c.Render(w, r, "partial.html", data)
    } else {
        c.Render(w, r, "full-page.html", data)
    }
}
```

## Example: Complete Tour Modal

See the workbench project's `tour-modal.html` for a production example of:
- Multi-step navigation with Hyperscript
- Dynamic button text changes
- State management via data attributes
- Form submission with HTMX
- Keyboard navigation support
- Clean separation of concerns