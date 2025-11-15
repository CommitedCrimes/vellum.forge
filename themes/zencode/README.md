# ZenCode Theme

A dark-first, minimalist blog theme inspired by [opencode.ai](https://opencode.ai), featuring a modern monochromatic design with clean typography and generous whitespace.

## Theme Inheritance

ZenCode uses a **fallback system** that automatically inherits from the default theme. If a template file doesn't exist in the ZenCode theme, the system will automatically use the corresponding file from the `default` theme.

### How it Works

When loading a template:
1. **First**, check if the file exists in `themes/zencode/`
2. **If not found**, automatically load from `themes/default/`
3. **If still not found**, return an error

This means you only need to override the files you want to customize, making theme development much simpler!

### Example

If you delete `themes/zencode/pages/home.jet`, the application will automatically use `themes/default/pages/home.jet` instead. This allows you to:
- Start with a minimal theme and gradually customize
- Focus only on the pages you want to redesign
- Maintain consistency with base functionality

## Design Philosophy

ZenCode embraces simplicity and clarity through:

- **Dark-first aesthetic** - Optimized for low-light reading with elegant light mode support
- **Monochromatic palette** - Professional grayscale color scheme
- **Generous whitespace** - Breathing room that emphasizes content over clutter
- **Clean typography** - Bold headlines with clear hierarchy using Inter font family
- **Minimal interactions** - Subtle hover effects and smooth transitions
- **Mobile-first** - Fully responsive design that works beautifully on all devices

## Features

### Theme Toggle
Built-in dark/light mode toggle with automatic localStorage persistence. Users can switch between themes with a single click.

### Modern Blog Cards
The blog index uses a card-based layout featuring:
- Optional cover images with subtle hover animations
- Clean metadata display (date, tags)
- Prominent article titles
- Brief descriptions
- Clear call-to-action links

### Minimalist Post Layout
Individual blog posts feature:
- Breadcrumb navigation
- Full-width cover image support
- Tag categorization
- Optimized reading experience with perfect line-height and font sizing
- Code syntax highlighting with Nordic theme
- Beautiful table styling
- Mermaid diagram support

## Installation

### Using Environment Variables

Set the `THEME` environment variable to use the ZenCode theme:

```bash
export THEME=zencode
```

Or add it to your `.env` file:

```env
THEME=zencode
```

### Using Command Line

```bash
THEME=zencode ./vellumforge
```

## Color Palette

### Dark Theme (Default)
- **Background Primary**: `#0a0a0a` - Deep black for main background
- **Background Secondary**: `#1a1a1a` - Slightly lighter for cards/elevated elements
- **Text Primary**: `#f1ecec` - Off-white for primary content
- **Text Secondary**: `#b7b1b1` - Gray for secondary content
- **Text Tertiary**: `#656363` - Subtle gray for metadata
- **Borders**: `#2a2a2a` - Subtle borders
- **Accent**: `#cfcecd` - Light gray for highlights

### Light Theme
- **Background Primary**: `#ffffff` - Pure white
- **Background Secondary**: `#f8f8f8` - Light gray
- **Text Primary**: `#211e1e` - Nearly black
- **Text Secondary**: `#656363` - Medium gray
- **Borders**: `#e5e5e5` - Light borders

## Typography

The theme uses **Inter** as the primary font family, a modern sans-serif typeface designed for screen readability.

### Heading Sizes
- **H1**: 3rem (48px) - Bold, prominent page titles
- **H2**: 2rem (32px) - Section headings
- **H3**: 1.5rem (24px) - Subsections
- **H4**: 1.25rem (20px) - Minor headings

### Body Text
- **Size**: 1rem (16px) base, 1.0625rem (17px) in article content
- **Line Height**: 1.7 for optimal readability
- **Color**: Secondary text color for reduced eye strain

## Customization

### Modifying Colors

Edit `themes/zencode/assets/css/theme.css` and update the CSS custom properties in the `:root[data-theme="dark"]` and `:root[data-theme="light"]` selectors.

### Changing Fonts

Replace the Google Fonts import in `themes/zencode/layout.jet`:

```html
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800;900&display=swap" rel="stylesheet">
```

Then update the font-family in `theme.css`:

```css
body {
    font-family: 'YourFont', -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}
```

### Adding Custom Pages

Create new page templates in `themes/zencode/pages/` following the existing structure. All pages should extend the base layout:

```jet
{{extends "../layout.jet"}}

{{block title()}}Your Page Title{{end}}

{{block main()}}
    <div class="your-content">
        <!-- Your HTML here -->
    </div>
{{end}}
```

## File Structure

```
themes/zencode/
├── README.md                      # This file
├── layout.jet                     # Base layout template
├── assets/
│   └── css/
│       └── theme.css             # Main stylesheet
├── pages/
│   └── blog/
│       ├── index.jet             # Blog listing page
│       └── post.jet              # Single blog post
└── partials/
    ├── nav.jet                   # Navigation component
    ├── footer.jet                # Footer component
    └── head.jet                  # Additional head elements
```

## Browser Support

ZenCode supports all modern browsers:
- Chrome/Edge (latest)
- Firefox (latest)
- Safari (latest)
- Mobile browsers

The theme uses modern CSS features including:
- CSS Custom Properties (variables)
- CSS Grid
- Flexbox
- CSS Transitions

## Performance

The theme is optimized for performance:
- Minimal CSS (~500 lines)
- No external dependencies beyond Google Fonts
- Efficient selectors
- Hardware-accelerated transitions
- Lazy-loaded images

## Accessibility

ZenCode follows web accessibility best practices:
- Semantic HTML structure
- ARIA labels for interactive elements
- Sufficient color contrast ratios (WCAG AA compliant)
- Keyboard navigation support
- Responsive text sizing

## Credits

Design inspiration from [opencode.ai](https://opencode.ai) - a beautiful example of minimalist, developer-focused design.

## License

This theme is part of the VellumForge project and follows the same license.

---

**Enjoy your new minimalist blog theme!**
