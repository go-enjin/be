# Go-Enjin Theme System

The Go-Enjin theming system is inspired by the hugo theme system. A hugo
theme is not intended to be directly usable within Go-Enjin though that
would be a "nice to have" goal.

The Go-Enjin theming system provides the following features:

- include a theme from a local filesystem
- include a theme from an embedded filesystem
- include multiple themes
- one theme can extend one other theme
- easily switch between themes
- contains an embedded default pair of themes: dark and light
- caches the processing output when given the same inputs
- directory and file structure mimic hugo themes
- can be used as a standalone Go library

## Default Theme Structure:

```
    theme-name/
        layouts/
            _default/
                single.html
                list.html
            partials/
                footer.html
                footer/
                    nav.html
                    fineprint.html
                head.html
                head/
                    favicon.html
                    metadata.html
                    social.html
                header.html
                header/
                    nav.html
                    banner.html
    	static/
            styles.css
        archetypes/
            default.md
```