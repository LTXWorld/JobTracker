# JobView Android UI/UX 设计规范

## 1. 执行概要

本文档定义了 JobView Android 应用的完整 UI/UX 设计规范，基于 Material Design 3 指导原则，结合移动端用户体验最佳实践，提供统一、直观、高效的用户界面设计标准和实施指南。

## 2. 设计系统概览

### 2.1 设计原则

**简洁高效**
- 界面布局简洁明了，减少认知负荷
- 核心功能优先，次要功能合理隐藏
- 操作流程精简，最多3步完成主要任务

**一致性**
- 统一的视觉语言和交互模式
- 跨屏幕的导航一致性
- 组件行为的可预测性

**可访问性**
- 支持大字体和高对比度模式
- 完整的语音辅助支持
- 适应不同使用能力的用户

**响应式适配**
- 适配4.7"-12"屏幕尺寸
- 支持横屏和竖屏显示
- 折叠屏设备优化

### 2.2 Material Design 3 适配策略

```typescript
// 设计系统配置
const JobViewDesignSystem = {
  // Material Design 3 基础
  materialDesign: {
    version: '3.0',
    colorSystem: 'Dynamic Color',
    typography: 'Material Typography',
    components: 'Material Components',
  },

  // 移动端适配
  mobileOptimizations: {
    touchTargets: '48dp minimum',
    thumbReach: 'Bottom navigation priority',
    onehanded: '70% functions accessible',
    hapticFeedback: 'Contextual vibration',
  },

  // JobView 特定扩展
  customizations: {
    brandColors: 'Professional blue palette',
    statusColors: 'Semantic status indicators',
    dataVisualization: 'Charts and graphs',
    offlineStates: 'Clear offline indicators',
  }
}
```

## 3. 视觉设计系统

### 3.1 色彩系统

#### 3.1.1 主色调配置

```typescript
// 基于 Material Design 3 的色彩令牌
export const ColorTokens = {
  // 主要色彩
  primary: {
    0: '#000000',
    10: '#002022',
    20: '#003638',
    30: '#00504f',
    40: '#006b66',
    50: '#00867f',
    60: '#00a299',
    70: '#00beb4',
    80: '#00dccf',
    90: '#00fbed',
    95: '#7efff6',
    99: '#f4fffd',
    100: '#ffffff',
  },

  // 次要色彩
  secondary: {
    0: '#000000',
    10: '#0f1415',
    20: '#252a2b',
    30: '#3c4142',
    40: '#53595a',
    50: '#6c7273',
    60: '#858c8c',
    70: '#9fa6a6',
    80: '#bac1c1',
    90: '#d6dddd',
    95: '#e4ebeb',
    99: '#f8fafa',
    100: '#ffffff',
  },

  // 语义色彩
  semantic: {
    error: {
      light: '#ba1a1a',
      dark: '#ffb4ab',
      container: '#ffdad6',
    },
    warning: {
      light: '#8b5000',
      dark: '#ffb951',
      container: '#ffddb3',
    },
    success: {
      light: '#146c2e',
      dark: '#7fdc8c',
      container: '#c2f0a5',
    },
    info: {
      light: '#1976d2',
      dark: '#90caf9',
      container: '#e3f2fd',
    },
  },

  // 状态色彩（求职流程特定）
  applicationStatus: {
    applied: '#2196f3',        // 蓝色 - 已投递
    screening: '#ff9800',      // 橙色 - 筛选中
    interview: '#9c27b0',      // 紫色 - 面试中
    offered: '#4caf50',        // 绿色 - 已获offer
    rejected: '#f44336',       // 红色 - 已拒绝
    withdrawn: '#757575',      // 灰色 - 已撤销
  }
}

// 主题适配
export const ThemeConfiguration = {
  light: {
    surface: ColorTokens.primary[99],
    onSurface: ColorTokens.primary[10],
    surfaceVariant: ColorTokens.secondary[90],
    onSurfaceVariant: ColorTokens.secondary[30],
    background: ColorTokens.primary[99],
    onBackground: ColorTokens.primary[10],
    elevation: {
      level0: ColorTokens.primary[99],  // 表面
      level1: '#f7f2fa',               // 卡片
      level2: '#f1ecf6',               // 浮动按钮
      level3: '#eae6f2',               // 底部导航
      level4: '#e4dfed',               // 应用栏
      level5: '#ddd8e9',               // 弹窗
    },
  },

  dark: {
    surface: ColorTokens.primary[10],
    onSurface: ColorTokens.primary[90],
    surfaceVariant: ColorTokens.secondary[30],
    onSurfaceVariant: ColorTokens.secondary[80],
    background: ColorTokens.primary[10],
    onBackground: ColorTokens.primary[90],
    elevation: {
      level0: ColorTokens.primary[10],  // 表面
      level1: '#1e2021',               // 卡片
      level2: '#24282a',               // 浮动按钮
      level3: '#2a3033',               // 底部导航
      level4: '#2e363c',               // 应用栏
      level5: '#323c44',               // 弹窗
    },
  }
}
```

#### 3.1.2 色彩使用规范

**主要色彩应用**
```typescript
// 色彩使用指南
export const ColorUsageGuide = {
  // 界面元素
  interface: {
    primaryButton: 'primary.40 (light) / primary.80 (dark)',
    secondaryButton: 'secondary.40 (light) / secondary.80 (dark)',
    accentElement: 'primary.60',
    dividers: 'surfaceVariant with 12% opacity',
    icons: 'onSurface with 60% opacity',
  },

  // 状态指示
  status: {
    active: ColorTokens.applicationStatus.applied,
    progress: ColorTokens.applicationStatus.interview,
    success: ColorTokens.applicationStatus.offered,
    warning: ColorTokens.semantic.warning.light,
    error: ColorTokens.applicationStatus.rejected,
  },

  // 数据可视化
  charts: {
    primary: [
      ColorTokens.primary[40],
      ColorTokens.primary[60],
      ColorTokens.primary[80],
    ],
    categorical: [
      '#1976d2', '#388e3c', '#f57c00',
      '#7b1fa2', '#d32f2f', '#00796b',
      '#5d4037', '#455a64', '#e91e63',
    ],
    sequential: [
      ColorTokens.primary[20],
      ColorTokens.primary[40],
      ColorTokens.primary[60],
      ColorTokens.primary[80],
    ],
  }
}
```

### 3.2 字体系统

#### 3.2.1 字体层级

```typescript
export const TypographyScale = {
  // Material Design 3 字体层级
  displayLarge: {
    fontFamily: 'Roboto',
    fontSize: 57,
    lineHeight: 64,
    letterSpacing: -0.25,
    fontWeight: '400',
    usage: '大标题、启动画面',
  },

  displayMedium: {
    fontFamily: 'Roboto',
    fontSize: 45,
    lineHeight: 52,
    letterSpacing: 0,
    fontWeight: '400',
    usage: '中标题、空状态',
  },

  displaySmall: {
    fontFamily: 'Roboto',
    fontSize: 36,
    lineHeight: 44,
    letterSpacing: 0,
    fontWeight: '400',
    usage: '小标题、对话框标题',
  },

  headlineLarge: {
    fontFamily: 'Roboto',
    fontSize: 32,
    lineHeight: 40,
    letterSpacing: 0,
    fontWeight: '400',
    usage: '页面标题',
  },

  headlineMedium: {
    fontFamily: 'Roboto',
    fontSize: 28,
    lineHeight: 36,
    letterSpacing: 0,
    fontWeight: '400',
    usage: '卡片标题、分组标题',
  },

  headlineSmall: {
    fontFamily: 'Roboto',
    fontSize: 24,
    lineHeight: 32,
    letterSpacing: 0,
    fontWeight: '400',
    usage: '子页面标题',
  },

  titleLarge: {
    fontFamily: 'Roboto',
    fontSize: 22,
    lineHeight: 28,
    letterSpacing: 0,
    fontWeight: '500',
    usage: '工具栏标题',
  },

  titleMedium: {
    fontFamily: 'Roboto',
    fontSize: 16,
    lineHeight: 24,
    letterSpacing: 0.15,
    fontWeight: '500',
    usage: '列表主标题、标签页',
  },

  titleSmall: {
    fontFamily: 'Roboto',
    fontSize: 14,
    lineHeight: 20,
    letterSpacing: 0.1,
    fontWeight: '500',
    usage: '列表副标题、按钮',
  },

  bodyLarge: {
    fontFamily: 'Roboto',
    fontSize: 16,
    lineHeight: 24,
    letterSpacing: 0.15,
    fontWeight: '400',
    usage: '正文内容、描述',
  },

  bodyMedium: {
    fontFamily: 'Roboto',
    fontSize: 14,
    lineHeight: 20,
    letterSpacing: 0.25,
    fontWeight: '400',
    usage: '次要内容、辅助信息',
  },

  bodySmall: {
    fontFamily: 'Roboto',
    fontSize: 12,
    lineHeight: 16,
    letterSpacing: 0.4,
    fontWeight: '400',
    usage: '说明文字、时间戳',
  },

  labelLarge: {
    fontFamily: 'Roboto',
    fontSize: 14,
    lineHeight: 20,
    letterSpacing: 0.1,
    fontWeight: '500',
    usage: '大按钮文字',
  },

  labelMedium: {
    fontFamily: 'Roboto',
    fontSize: 12,
    lineHeight: 16,
    letterSpacing: 0.5,
    fontWeight: '500',
    usage: '标准按钮文字、标签',
  },

  labelSmall: {
    fontFamily: 'Roboto',
    fontSize: 11,
    lineHeight: 16,
    letterSpacing: 0.5,
    fontWeight: '500',
    usage: '小按钮文字、徽章',
  },
}

// 中文字体适配
export const ChineseTypographyEnhancement = {
  fontFamily: {
    primary: 'Roboto, "Noto Sans CJK SC", "PingFang SC", "Hiragino Sans GB", sans-serif',
    display: 'Roboto, "Noto Sans CJK SC", "PingFang SC", sans-serif',
    monospace: '"Roboto Mono", "Noto Sans Mono CJK SC", monospace',
  },

  // 中文特殊调整
  adjustments: {
    lineHeightIncrease: 0.1,  // 中文需要更大的行高
    letterSpacingReduce: 0.05, // 中文字符间距适当减少
  },
}
```

#### 3.2.2 响应式字体

```typescript
export const ResponsiveTypography = {
  // 基于屏幕尺寸的字体缩放
  breakpoints: {
    small: { maxWidth: 360, scale: 0.9 },
    medium: { maxWidth: 768, scale: 1.0 },
    large: { maxWidth: 1024, scale: 1.1 },
  },

  // 可访问性字体设置
  accessibility: {
    minimumTouchTarget: 48, // dp
    maximumLineLength: 75,  // 字符
    minimumContrast: 4.5,   // WCAG AA 标准
    largeTextContrast: 3.0, // 大字体对比度
  },

  // 动态字体支持
  dynamicType: {
    scale: {
      small: 0.85,
      medium: 1.0,
      large: 1.15,
      xlarge: 1.3,
      xxlarge: 1.5,
    },
    maxScaleFactor: 2.0, // 最大放大倍数
  },
}
```

### 3.3 间距系统

#### 3.3.1 空间令牌

```typescript
export const SpacingTokens = {
  // 基础间距单位 (4dp)
  base: 4,

  // 间距量级
  spacing: {
    xs: 4,   // 极小间距 - 紧密相关元素
    sm: 8,   // 小间距 - 相关元素
    md: 16,  // 中间距 - 标准间距
    lg: 24,  // 大间距 - 分组间距
    xl: 32,  // 超大间距 - 区块间距
    xxl: 48, // 极大间距 - 页面分区
  },

  // 组件内间距
  inset: {
    xs: 4,   // 徽章、标签内间距
    sm: 8,   // 按钮内间距
    md: 16,  // 卡片内间距
    lg: 24,  // 页面边距
    xl: 32,  // 大容器边距
  },

  // 组件间距
  stack: {
    xs: 4,   // 表单字段间距
    sm: 8,   // 列表项间距
    md: 16,  // 卡片间距
    lg: 24,  // 区块间距
    xl: 32,  // 页面区块间距
  },

  // 内联间距
  inline: {
    xs: 4,   // 图标与文字
    sm: 8,   // 相关按钮
    md: 12,  // 工具栏按钮
    lg: 16,  // 导航项
    xl: 24,  // 主要操作区分
  },
}

// 响应式间距
export const ResponsiveSpacing = {
  // 基于屏幕尺寸调整
  small: { multiplier: 0.75 },  // 小屏幕间距减少25%
  medium: { multiplier: 1.0 },  // 标准屏幕
  large: { multiplier: 1.25 },  // 大屏幕间距增加25%

  // 安全区域适配
  safeArea: {
    top: 'env(safe-area-inset-top)',
    bottom: 'env(safe-area-inset-bottom)',
    left: 'env(safe-area-inset-left)',
    right: 'env(safe-area-inset-right)',
  },
}
```

#### 3.3.2 布局网格

```typescript
export const GridSystem = {
  // Material Design 网格
  columns: 4,  // 移动端4列网格
  gutter: 16,  // 列间距
  margin: 16,  // 页面边距

  // 断点适配
  breakpoints: {
    xs: { columns: 4, gutter: 8, margin: 16 },   // 320-599dp
    sm: { columns: 8, gutter: 16, margin: 16 },  // 600-1023dp
    md: { columns: 8, gutter: 24, margin: 24 },  // 1024-1439dp
    lg: { columns: 12, gutter: 24, margin: 24 }, // 1440dp+
  },

  // 组件布局
  componentLayouts: {
    card: {
      minWidth: 280,
      maxWidth: 400,
      aspectRatio: '16:9',
    },
    list: {
      minItemHeight: 56,
      maxItemHeight: 72,
    },
    form: {
      fieldHeight: 56,
      fieldSpacing: 16,
    },
  },
}
```

## 4. 组件设计规范

### 4.1 基础组件

#### 4.1.1 按钮组件

```typescript
// 按钮规范
export const ButtonSpecification = {
  // 按钮类型
  types: {
    filled: {
      background: 'primary.40',
      foreground: 'primary.100',
      elevation: 'level1',
      usage: '主要操作',
    },
    outlined: {
      background: 'transparent',
      foreground: 'primary.40',
      border: 'primary.40',
      usage: '次要操作',
    },
    text: {
      background: 'transparent',
      foreground: 'primary.40',
      usage: '辅助操作',
    },
    tonal: {
      background: 'primary.90',
      foreground: 'primary.10',
      usage: '中等重要度操作',
    },
  },

  // 尺寸规范
  sizes: {
    small: {
      height: 32,
      paddingHorizontal: 12,
      fontSize: TypographyScale.labelSmall.fontSize,
    },
    medium: {
      height: 40,
      paddingHorizontal: 16,
      fontSize: TypographyScale.labelMedium.fontSize,
    },
    large: {
      height: 48,
      paddingHorizontal: 24,
      fontSize: TypographyScale.labelLarge.fontSize,
    },
  },

  // 状态样式
  states: {
    default: { opacity: 1.0 },
    hovered: { opacity: 0.92, elevation: 'level2' },
    focused: { opacity: 0.92, outline: 'primary.40' },
    pressed: { opacity: 0.88, elevation: 'level0' },
    disabled: { opacity: 0.38 },
  },

  // 交互规范
  interactions: {
    rippleEffect: true,
    hapticFeedback: 'light',
    animationDuration: 150, // ms
    minimumTouchTarget: 48, // dp
  },
}

// React Native 实现示例
const ButtonComponent: React.FC<ButtonProps> = ({
  type = 'filled',
  size = 'medium',
  disabled = false,
  children,
  onPress,
  ...props
}) => {
  const theme = useTheme()
  const buttonConfig = ButtonSpecification.types[type]
  const sizeConfig = ButtonSpecification.sizes[size]

  return (
    <TouchableOpacity
      style={[
        styles.button,
        {
          height: sizeConfig.height,
          paddingHorizontal: sizeConfig.paddingHorizontal,
          backgroundColor: theme.colors[buttonConfig.background],
          borderColor: buttonConfig.border ? theme.colors[buttonConfig.border] : undefined,
          opacity: disabled ? ButtonSpecification.states.disabled.opacity : 1,
        }
      ]}
      onPress={disabled ? undefined : onPress}
      disabled={disabled}
      activeOpacity={ButtonSpecification.states.pressed.opacity}
      {...props}
    >
      <Text
        style={{
          color: theme.colors[buttonConfig.foreground],
          fontSize: sizeConfig.fontSize,
          fontWeight: '500',
        }}
      >
        {children}
      </Text>
    </TouchableOpacity>
  )
}
```

#### 4.1.2 卡片组件

```typescript
export const CardSpecification = {
  // 卡片类型
  types: {
    elevated: {
      elevation: 'level1',
      background: 'surface',
      usage: '标准信息卡片',
    },
    filled: {
      background: 'surfaceVariant',
      usage: '次要信息卡片',
    },
    outlined: {
      background: 'surface',
      border: 'outline',
      usage: '强调边界的卡片',
    },
  },

  // 布局规范
  layout: {
    borderRadius: 12,
    padding: SpacingTokens.inset.md,
    minHeight: 72,
    elevation: {
      level0: 0,
      level1: 2,
      level2: 4,
      level3: 8,
    },
  },

  // 内容布局
  content: {
    header: {
      marginBottom: SpacingTokens.stack.sm,
    },
    media: {
      aspectRatio: '16:9',
      borderRadius: 8,
    },
    actions: {
      marginTop: SpacingTokens.stack.md,
      gap: SpacingTokens.inline.sm,
    },
  },
}

// 投递记录卡片特殊规范
export const ApplicationCardSpecification = {
  layout: {
    ...CardSpecification.layout,
    minHeight: 120,
  },

  // 状态指示器
  statusIndicator: {
    width: 4,
    height: '100%',
    position: 'absolute',
    left: 0,
    top: 0,
    borderRadius: 2,
  },

  // 内容区域
  contentArea: {
    paddingLeft: SpacingTokens.inset.md + 8, // 为状态指示器留空间
  },

  // 操作区域
  actions: {
    position: 'absolute',
    right: SpacingTokens.inset.sm,
    top: SpacingTokens.inset.sm,
  },
}
```

#### 4.1.3 输入组件

```typescript
export const InputSpecification = {
  // 输入框类型
  types: {
    filled: {
      background: 'surfaceVariant',
      bottomBorder: 'outline',
      usage: '标准输入',
    },
    outlined: {
      background: 'transparent',
      border: 'outline',
      borderRadius: 4,
      usage: '突出显示的输入',
    },
  },

  // 尺寸规范
  sizes: {
    medium: {
      height: 56,
      fontSize: TypographyScale.bodyLarge.fontSize,
      padding: SpacingTokens.inset.md,
    },
    large: {
      height: 64,
      fontSize: TypographyScale.bodyLarge.fontSize,
      padding: SpacingTokens.inset.lg,
    },
  },

  // 状态样式
  states: {
    default: {
      borderColor: 'outline',
      labelColor: 'onSurfaceVariant',
    },
    focused: {
      borderColor: 'primary.40',
      borderWidth: 2,
      labelColor: 'primary.40',
    },
    error: {
      borderColor: 'error.40',
      labelColor: 'error.40',
    },
    disabled: {
      opacity: 0.38,
      backgroundColor: 'surface',
    },
  },

  // 辅助元素
  supporting: {
    label: {
      typography: TypographyScale.bodySmall,
      marginBottom: SpacingTokens.stack.xs,
    },
    helperText: {
      typography: TypographyScale.bodySmall,
      marginTop: SpacingTokens.stack.xs,
    },
    errorText: {
      typography: TypographyScale.bodySmall,
      color: 'error.40',
      marginTop: SpacingTokens.stack.xs,
    },
  },
}
```

### 4.2 复合组件

#### 4.2.1 导航组件

```typescript
export const NavigationSpecification = {
  // 底部导航
  bottomNavigation: {
    height: 80,
    background: 'surface',
    elevation: 'level2',

    item: {
      width: '25%', // 4个标签页
      height: 64,
      padding: SpacingTokens.inset.xs,
      gap: SpacingTokens.stack.xs,
    },

    icon: {
      size: 24,
      activeColor: 'primary.40',
      inactiveColor: 'onSurfaceVariant',
    },

    label: {
      typography: TypographyScale.labelSmall,
      activeColor: 'primary.40',
      inactiveColor: 'onSurfaceVariant',
    },

    indicator: {
      width: 64,
      height: 32,
      borderRadius: 16,
      background: 'primary.90',
    },
  },

  // 顶部应用栏
  topAppBar: {
    small: {
      height: 64,
      padding: SpacingTokens.inset.md,
    },
    medium: {
      height: 112,
      padding: SpacingTokens.inset.md,
    },
    large: {
      height: 152,
      padding: SpacingTokens.inset.md,
    },

    title: {
      typography: TypographyScale.titleLarge,
      color: 'onSurface',
    },

    actions: {
      gap: SpacingTokens.inline.sm,
    },
  },

  // 侧边抽屉
  drawer: {
    width: 280,
    maxWidth: '80%',
    background: 'surface',
    elevation: 'level1',

    header: {
      height: 164,
      padding: SpacingTokens.inset.lg,
      background: 'primary.90',
    },

    item: {
      height: 56,
      padding: SpacingTokens.inset.md,
      borderRadius: 28,
      margin: `0 ${SpacingTokens.inset.sm}px`,
    },
  },
}
```

#### 4.2.2 列表组件

```typescript
export const ListSpecification = {
  // 列表项类型
  items: {
    oneLineItem: {
      height: 56,
      padding: SpacingTokens.inset.md,

      leading: {
        width: 24,
        height: 24,
        marginRight: SpacingTokens.inline.md,
      },

      text: {
        typography: TypographyScale.bodyLarge,
        color: 'onSurface',
      },

      trailing: {
        marginLeft: 'auto',
      },
    },

    twoLineItem: {
      height: 72,
      padding: SpacingTokens.inset.md,

      text: {
        primary: {
          typography: TypographyScale.bodyLarge,
          color: 'onSurface',
        },
        secondary: {
          typography: TypographyScale.bodyMedium,
          color: 'onSurfaceVariant',
          marginTop: SpacingTokens.stack.xs,
        },
      },
    },

    threeLineItem: {
      height: 88,
      padding: SpacingTokens.inset.md,

      text: {
        primary: {
          typography: TypographyScale.bodyLarge,
          color: 'onSurface',
        },
        secondary: {
          typography: TypographyScale.bodyMedium,
          color: 'onSurfaceVariant',
          marginTop: SpacingTokens.stack.xs,
          numberOfLines: 2,
        },
      },
    },
  },

  // 分组和分隔
  dividers: {
    full: {
      height: 1,
      background: 'outline',
      opacity: 0.12,
    },
    inset: {
      height: 1,
      background: 'outline',
      opacity: 0.12,
      marginLeft: 56, // 考虑图标空间
    },
  },

  sections: {
    header: {
      height: 48,
      padding: SpacingTokens.inset.md,
      typography: TypographyScale.titleMedium,
      color: 'primary.40',
    },

    spacing: SpacingTokens.stack.lg,
  },
}

// 投递记录列表特殊规范
export const ApplicationListSpecification = {
  item: {
    ...ListSpecification.items.twoLineItem,
    height: 88, // 增加高度以显示更多信息
  },

  statusBadge: {
    height: 20,
    paddingHorizontal: 8,
    borderRadius: 10,
    marginTop: SpacingTokens.stack.xs,
  },

  dateInfo: {
    typography: TypographyScale.bodySmall,
    color: 'onSurfaceVariant',
    marginTop: SpacingTokens.stack.xs,
  },

  quickActions: {
    width: 80,
    flexDirection: 'row',
    gap: SpacingTokens.inline.xs,
  },
}
```

## 5. 移动端交互设计

### 5.1 手势交互

```typescript
export const GestureSpecification = {
  // 基础手势
  tap: {
    minimumTouchTarget: 48, // dp
    rippleEffect: true,
    hapticFeedback: 'light',
    timeout: 300, // ms
  },

  longPress: {
    minimumDuration: 500, // ms
    hapticFeedback: 'medium',
    contextMenu: true,
  },

  swipe: {
    threshold: 50, // dp
    velocity: 300, // dp/s
    directions: ['left', 'right', 'up', 'down'],
  },

  pinch: {
    minimumScale: 0.5,
    maximumScale: 3.0,
    sensitivity: 1.0,
  },

  // 应用特定手势
  applicationSpecific: {
    // 卡片滑动操作
    cardSwipe: {
      leftSwipe: {
        action: 'markComplete',
        icon: 'check',
        color: ColorTokens.semantic.success.light,
        threshold: 80,
      },
      rightSwipe: {
        action: 'archive',
        icon: 'archive',
        color: ColorTokens.secondary[40],
        threshold: 80,
      },
    },

    // 列表项长按
    listItemLongPress: {
      duration: 600,
      actions: ['edit', 'delete', 'share'],
      haptic: 'medium',
    },

    // 下拉刷新
    pullToRefresh: {
      threshold: 100,
      maxDistance: 150,
      hapticFeedback: 'light',
      animation: 'spring',
    },
  },
}

// 单手操作优化
export const OneHandedOperationOptimization = {
  // 拇指可触及区域
  thumbReachZones: {
    easy: {
      description: '拇指轻松触及',
      bounds: { bottom: '60%', left: 0, right: '100%' },
      usage: '主要操作按钮',
    },
    stretch: {
      description: '拇指伸展可及',
      bounds: { top: '40%', bottom: '75%', left: 0, right: '100%' },
      usage: '次要操作',
    },
    difficult: {
      description: '双手或换手操作',
      bounds: { top: 0, bottom: '40%', left: 0, right: '100%' },
      usage: '信息显示、设置',
    },
  },

  // 操作优先级布局
  actionPlacement: {
    primary: 'bottom-right', // 主操作按钮
    secondary: 'bottom-center', // 次要操作
    navigation: 'bottom', // 导航栏
    menu: 'top-left', // 菜单按钮
    search: 'top-right', // 搜索按钮
  },

  // 可达性增强
  reachabilityEnhancement: {
    floatingActionButton: {
      position: { bottom: 16, right: 16 },
      size: 56,
      elevation: 'level3',
    },
    quickActions: {
      position: 'bottom',
      height: 56,
      background: 'surface',
    },
  },
}
```

### 5.2 动画和过渡

```typescript
export const AnimationSpecification = {
  // 动画时长
  duration: {
    micro: 50,    // 状态变化
    short: 150,   // 小组件动画
    medium: 250,  // 页面元素动画
    long: 500,    // 页面过渡
    extra: 1000,  // 复杂动画
  },

  // 缓动函数
  easing: {
    standard: 'cubic-bezier(0.4, 0.0, 0.2, 1)',     // 标准缓动
    decelerate: 'cubic-bezier(0.0, 0.0, 0.2, 1)',  // 减速
    accelerate: 'cubic-bezier(0.4, 0.0, 1, 1)',    // 加速
    sharp: 'cubic-bezier(0.4, 0.0, 0.6, 1)',       // 锐利
  },

  // 页面过渡
  pageTransitions: {
    slideFromRight: {
      enter: {
        transform: [{ translateX: 300 }],
        opacity: 0,
      },
      enterActive: {
        transform: [{ translateX: 0 }],
        opacity: 1,
        duration: AnimationSpecification.duration.medium,
      },
    },

    slideFromBottom: {
      enter: {
        transform: [{ translateY: 300 }],
        opacity: 0,
      },
      enterActive: {
        transform: [{ translateY: 0 }],
        opacity: 1,
        duration: AnimationSpecification.duration.medium,
      },
    },

    fade: {
      enter: { opacity: 0 },
      enterActive: {
        opacity: 1,
        duration: AnimationSpecification.duration.short,
      },
    },
  },

  // 组件动画
  componentAnimations: {
    buttonPress: {
      scale: 0.95,
      duration: AnimationSpecification.duration.micro,
    },

    cardElevation: {
      elevation: 'level1 -> level2',
      duration: AnimationSpecification.duration.short,
    },

    listItemReveal: {
      opacity: '0 -> 1',
      transform: [{ translateY: 20 }, { translateY: 0 }],
      duration: AnimationSpecification.duration.short,
      delay: (index) => index * 50, // 交错动画
    },

    modalPresentation: {
      backdrop: {
        opacity: '0 -> 0.6',
        duration: AnimationSpecification.duration.medium,
      },
      modal: {
        transform: [
          { scale: 0.9 }, { scale: 1.0 },
          { translateY: 50 }, { translateY: 0 }
        ],
        opacity: '0 -> 1',
        duration: AnimationSpecification.duration.medium,
      },
    },
  },

  // 微交互动画
  microInteractions: {
    rippleEffect: {
      scale: '0 -> 1',
      opacity: '0.12 -> 0',
      duration: AnimationSpecification.duration.medium,
    },

    checkboxCheck: {
      strokeDasharray: '0 -> 20',
      duration: AnimationSpecification.duration.short,
    },

    switchToggle: {
      thumbPosition: 'left -> right',
      trackColor: 'outline -> primary',
      duration: AnimationSpecification.duration.short,
    },
  },
}
```

## 6. 响应式设计

### 6.1 屏幕尺寸适配

```typescript
export const ResponsiveDesignSystem = {
  // 屏幕尺寸断点
  breakpoints: {
    xs: { maxWidth: 360, name: 'small phone' },
    sm: { maxWidth: 640, name: 'phone' },
    md: { maxWidth: 768, name: 'tablet' },
    lg: { maxWidth: 1024, name: 'large tablet' },
    xl: { maxWidth: 1440, name: 'desktop' },
  },

  // 布局适配
  layouts: {
    singlePane: {
      breakpoints: ['xs', 'sm'],
      navigation: 'bottom-tabs',
      content: 'single-column',
    },

    dualPane: {
      breakpoints: ['md', 'lg'],
      navigation: 'side-rail',
      content: 'master-detail',
      ratio: '1:2', // 主列表:详情
    },

    multiPane: {
      breakpoints: ['xl'],
      navigation: 'permanent-drawer',
      content: 'three-column',
      ratio: '1:2:1', // 导航:主内容:辅助
    },
  },

  // 组件适配
  componentAdaptations: {
    applicationCard: {
      xs: { width: '100%', margin: 8 },
      sm: { width: '100%', margin: 16 },
      md: { width: '48%', margin: 16 },
      lg: { width: '32%', margin: 24 },
    },

    bottomSheet: {
      xs: { height: '90%' },
      sm: { height: '80%' },
      md: { height: '70%', width: 640, centered: true },
      lg: { height: '60%', width: 800, centered: true },
    },

    dataTable: {
      xs: { hidden: true, useCardLayout: true },
      sm: { columns: 3, scrollable: true },
      md: { columns: 5, scrollable: false },
      lg: { columns: 7, scrollable: false },
    },
  },
}

// 密度适配
export const DensityAdaptation = {
  // 基于像素密度调整
  densities: {
    ldpi: { scale: 0.75, suffix: '' },      // ~120dpi
    mdpi: { scale: 1.0, suffix: '' },       // ~160dpi
    hdpi: { scale: 1.5, suffix: '@2x' },    // ~240dpi
    xhdpi: { scale: 2.0, suffix: '@3x' },   // ~320dpi
    xxhdpi: { scale: 3.0, suffix: '@4x' },  // ~480dpi
  },

  // 图标尺寸适配
  iconSizes: {
    small: { base: 16, scale: true },
    medium: { base: 24, scale: true },
    large: { base: 32, scale: true },
    xlarge: { base: 48, scale: true },
  },

  // 触摸目标适配
  touchTargets: {
    minimum: 48, // dp
    recommended: 56, // dp
    comfortable: 64, // dp
  },
}
```

### 6.2 方向适配

```typescript
export const OrientationAdaptation = {
  // 纵向布局
  portrait: {
    navigation: 'bottom-tabs',
    appBar: 'small',
    content: {
      maxWidth: '100%',
      padding: SpacingTokens.inset.md,
    },

    kanbanBoard: {
      columns: 2, // 2列状态
      scrollDirection: 'vertical',
    },

    chartArea: {
      height: 300,
      aspectRatio: '16:9',
    },
  },

  // 横向布局
  landscape: {
    navigation: 'side-rail',
    appBar: 'hidden-on-scroll',
    content: {
      maxWidth: '100%',
      padding: SpacingTokens.inset.lg,
    },

    kanbanBoard: {
      columns: 4, // 4列状态
      scrollDirection: 'horizontal',
    },

    chartArea: {
      height: 200,
      aspectRatio: '21:9',
    },
  },

  // 方向切换动画
  orientationTransition: {
    duration: AnimationSpecification.duration.medium,
    easing: AnimationSpecification.easing.standard,
  },
}
```

## 7. 可访问性设计

### 7.1 无障碍标准

```typescript
export const AccessibilitySpecification = {
  // WCAG 2.1 AA 标准
  wcagCompliance: {
    level: 'AA',
    guidelines: {
      perceivable: true,
      operable: true,
      understandable: true,
      robust: true,
    },
  },

  // 颜色对比度
  colorContrast: {
    normalText: 4.5,      // AA 标准
    largeText: 3.0,       // 18pt+ 或 14pt+ 粗体
    uiComponents: 3.0,    // 交互组件
    graphicalObjects: 3.0, // 图标和图形
  },

  // 字体大小
  fontSize: {
    minimum: 12,          // sp
    bodyText: 14,         // sp
    headings: 18,         // sp
    maximumScale: 2.0,    // 用户可放大倍数
  },

  // 触摸目标
  touchTargets: {
    minimum: 44,          // dp (iOS标准)
    recommended: 48,      // dp (Android标准)
    spacing: 8,           // dp 目标间最小间距
  },

  // 语义化标签
  semanticLabels: {
    button: {
      role: 'button',
      accessibilityLabel: 'required',
      accessibilityHint: 'optional',
    },

    textInput: {
      role: 'textbox',
      accessibilityLabel: 'required',
      accessibilityValue: 'current value',
    },

    image: {
      role: 'image',
      accessibilityLabel: 'required',
      decorative: 'alt=""',
    },

    navigation: {
      role: 'navigation',
      accessibilityLabel: 'required',
      currentPage: 'aria-current="page"',
    },
  },
}

// 屏幕阅读器支持
export const ScreenReaderSupport = {
  // VoiceOver (iOS) / TalkBack (Android)
  announcements: {
    statusChange: {
      priority: 'assertive',
      message: (status) => `状态已更改为${status}`,
    },

    dataUpdate: {
      priority: 'polite',
      message: (count) => `已加载${count}条新记录`,
    },

    error: {
      priority: 'assertive',
      message: (error) => `错误：${error}`,
    },
  },

  // 导航顺序
  navigationOrder: {
    tabOrder: 'logical',
    skipLinks: true,
    landmarks: ['header', 'main', 'navigation', 'footer'],
  },

  // 内容结构
  contentStructure: {
    headingLevels: 'hierarchical',
    lists: 'semantic-markup',
    tables: 'headers-associated',
    forms: 'labeled-controls',
  },
}
```

### 7.2 动效减少支持

```typescript
export const MotionAccessibility = {
  // 减少动效设置
  reducedMotion: {
    detection: 'prefers-reduced-motion: reduce',

    alternatives: {
      fadeIn: { replace: 'instant-appear' },
      slideTransition: { replace: 'fade-transition' },
      bounce: { replace: 'linear-scale' },
      parallax: { replace: 'static-background' },
    },

    exceptions: {
      loading: 'allow', // 加载动画保持
      focus: 'allow',   // 焦点指示保持
      error: 'allow',   // 错误提示保持
    },
  },

  // 动画控制
  animationControls: {
    globalSwitch: true,
    categoryControls: {
      pageTransitions: true,
      uiAnimations: true,
      decorativeAnimations: true,
    },

    playbackControls: {
      pause: true,
      speed: [0.5, 1.0, 1.5, 2.0],
    },
  },
}
```

## 8. 主题和个性化

### 8.1 动态主题系统

```typescript
export const DynamicThemeSystem = {
  // Material You 动态颜色
  dynamicColor: {
    enabled: true,
    source: 'wallpaper', // 从壁纸提取颜色
    fallback: ColorTokens.primary, // 降级方案

    extraction: {
      algorithm: 'quantize',
      colorCount: 16,
      quality: 1,
    },

    generation: {
      lightTheme: 'auto-generate',
      darkTheme: 'auto-generate',
      contrastTheme: 'auto-generate',
    },
  },

  // 预设主题
  presetThemes: {
    jobViewBlue: {
      primary: '#1976d2',
      secondary: '#424242',
      accent: '#03a9f4',
      name: 'JobView 蓝色',
    },

    professionalGreen: {
      primary: '#2e7d4f',
      secondary: '#545454',
      accent: '#4caf50',
      name: '专业绿色',
    },

    warmOrange: {
      primary: '#f57c00',
      secondary: '#5d4037',
      accent: '#ff9800',
      name: '温暖橙色',
    },

    elegantPurple: {
      primary: '#7b1fa2',
      secondary: '#4a148c',
      accent: '#9c27b0',
      name: '优雅紫色',
    },
  },

  // 主题切换
  switching: {
    method: 'smooth-transition',
    duration: 300,
    easing: 'ease-in-out',
    preserveImages: true,
  },
}

// 个性化设置
export const PersonalizationOptions = {
  // 视觉定制
  visual: {
    // 卡片显示样式
    cardStyle: {
      options: ['compact', 'comfortable', 'spacious'],
      default: 'comfortable',
    },

    // 状态显示方式
    statusDisplay: {
      options: ['color-bar', 'badge', 'background'],
      default: 'color-bar',
    },

    // 列表密度
    listDensity: {
      options: ['compact', 'standard', 'comfortable'],
      default: 'standard',
      affects: ['itemHeight', 'padding', 'fontSize'],
    },

    // 字体大小
    fontScale: {
      range: [0.85, 1.3],
      step: 0.05,
      default: 1.0,
    },
  },

  // 功能定制
  functional: {
    // 默认视图
    defaultView: {
      options: ['kanban', 'timeline', 'list'],
      default: 'kanban',
    },

    // 快速操作
    quickActions: {
      customizable: true,
      maxCount: 4,
      availableActions: [
        'updateStatus',
        'addNote',
        'setReminder',
        'share',
        'duplicate',
        'archive',
      ],
    },

    // 筛选预设
    filterPresets: {
      customizable: true,
      maxCount: 5,
      shareable: true,
    },
  },

  // 行为设置
  behavioral: {
    // 手势操作
    gestureActions: {
      swipeLeft: {
        options: ['markComplete', 'archive', 'delete', 'none'],
        default: 'markComplete',
      },
      swipeRight: {
        options: ['edit', 'duplicate', 'share', 'none'],
        default: 'edit',
      },
      longPress: {
        options: ['contextMenu', 'quickEdit', 'none'],
        default: 'contextMenu',
      },
    },

    // 自动操作
    autoActions: {
      autoSync: {
        enabled: true,
        frequency: 'real-time',
      },
      autoBackup: {
        enabled: true,
        frequency: 'daily',
      },
    },
  },
}
```

### 8.2 暗色主题设计

```typescript
export const DarkThemeSpecification = {
  // 暗色主题原则
  principles: {
    elevation: 'lighter-surfaces-higher',
    contrast: 'maintain-accessibility',
    colorUsage: 'desaturated-primaries',
    content: 'primary-content-emphasis',
  },

  // 表面颜色
  surfaces: {
    level0: '#121212',  // 背景
    level1: '#1e1e1e',  // 卡片
    level2: '#232323',  // 浮动按钮
    level3: '#252525',  // 抽屉
    level4: '#272727',  // 应用栏
    level5: '#2c2c2c',  // 对话框
  },

  // 内容颜色
  content: {
    highEmphasis: 'rgba(255, 255, 255, 0.87)',
    mediumEmphasis: 'rgba(255, 255, 255, 0.60)',
    disabled: 'rgba(255, 255, 255, 0.38)',
  },

  // 特殊处理
  specialCases: {
    // 图片和媒体
    images: {
      overlay: 'rgba(0, 0, 0, 0.3)',
      brightnessReduction: 0.8,
    },

    // 图表颜色
    charts: {
      background: 'transparent',
      gridLines: 'rgba(255, 255, 255, 0.12)',
      text: 'rgba(255, 255, 255, 0.87)',
    },

    // 状态颜色（去饱和处理）
    statusColors: {
      applied: '#1565c0',
      screening: '#e65100',
      interview: '#6a1b9a',
      offered: '#2e7d4f',
      rejected: '#c62828',
    },
  },
}
```

## 9. 开发实施指南

### 9.1 组件库结构

```typescript
// 设计系统组件库结构
export const ComponentLibraryStructure = {
  // 基础令牌
  tokens: {
    colors: 'ColorTokens',
    typography: 'TypographyScale',
    spacing: 'SpacingTokens',
    elevation: 'ElevationTokens',
    motion: 'AnimationSpecification',
  },

  // 基础组件
  primitives: {
    Button: 'ButtonComponent',
    Input: 'InputComponent',
    Card: 'CardComponent',
    List: 'ListComponent',
    Modal: 'ModalComponent',
  },

  // 复合组件
  composite: {
    ApplicationCard: 'ApplicationCardComponent',
    StatusBadge: 'StatusBadgeComponent',
    KanbanColumn: 'KanbanColumnComponent',
    StatisticsChart: 'ChartComponent',
  },

  // 布局组件
  layout: {
    Screen: 'ScreenComponent',
    Section: 'SectionComponent',
    Grid: 'GridComponent',
    Stack: 'StackComponent',
  },

  // 导航组件
  navigation: {
    BottomTabs: 'BottomTabsComponent',
    TopAppBar: 'TopAppBarComponent',
    Drawer: 'DrawerComponent',
  },
}
```

### 9.2 设计令牌实现

```typescript
// React Native 样式实现
import { StyleSheet, Platform, Dimensions } from 'react-native'

export const createStyles = (theme: Theme) => {
  const { colors, typography, spacing } = theme

  return StyleSheet.create({
    // 基础样式
    container: {
      flex: 1,
      backgroundColor: colors.background,
      padding: spacing.inset.md,
    },

    card: {
      backgroundColor: colors.surface,
      borderRadius: 12,
      padding: spacing.inset.md,
      marginBottom: spacing.stack.md,

      // 平台特定阴影
      ...Platform.select({
        ios: {
          shadowColor: colors.shadow,
          shadowOffset: { width: 0, height: 2 },
          shadowOpacity: 0.1,
          shadowRadius: 4,
        },
        android: {
          elevation: 2,
        },
      }),
    },

    // 文字样式
    headlineText: {
      fontFamily: typography.headlineMedium.fontFamily,
      fontSize: typography.headlineMedium.fontSize,
      lineHeight: typography.headlineMedium.lineHeight,
      fontWeight: typography.headlineMedium.fontWeight,
      color: colors.onSurface,
    },

    bodyText: {
      fontFamily: typography.bodyLarge.fontFamily,
      fontSize: typography.bodyLarge.fontSize,
      lineHeight: typography.bodyLarge.lineHeight,
      color: colors.onSurface,
    },

    // 按钮样式
    primaryButton: {
      backgroundColor: colors.primary,
      paddingHorizontal: spacing.inset.lg,
      paddingVertical: spacing.inset.md,
      borderRadius: 24,
      minHeight: 48,
      justifyContent: 'center',
      alignItems: 'center',
    },

    primaryButtonText: {
      color: colors.onPrimary,
      fontSize: typography.labelLarge.fontSize,
      fontWeight: typography.labelLarge.fontWeight,
    },
  })
}

// 响应式工具函数
export const useResponsiveStyles = () => {
  const { width, height } = Dimensions.get('window')

  const isSmallScreen = width < 360
  const isTablet = width >= 768
  const isLandscape = width > height

  return {
    isSmallScreen,
    isTablet,
    isLandscape,
    screenWidth: width,
    screenHeight: height,
  }
}
```

### 9.3 主题切换实现

```typescript
// 主题上下文
export const ThemeContext = React.createContext<{
  theme: Theme
  isDark: boolean
  toggleTheme: () => void
  setTheme: (themeName: string) => void
}>({
  theme: lightTheme,
  isDark: false,
  toggleTheme: () => {},
  setTheme: () => {},
})

// 主题提供者
export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [currentTheme, setCurrentTheme] = useState<string>('light')
  const [customColors, setCustomColors] = useState<any>(null)

  // 监听系统主题变化
  useEffect(() => {
    const subscription = Appearance.addChangeListener(({ colorScheme }) => {
      if (currentTheme === 'system') {
        setCurrentTheme(colorScheme || 'light')
      }
    })

    return () => subscription?.remove()
  }, [currentTheme])

  // 生成当前主题
  const theme = useMemo(() => {
    const baseTheme = currentTheme === 'dark' ? darkTheme : lightTheme

    if (customColors) {
      return generateCustomTheme(baseTheme, customColors)
    }

    return baseTheme
  }, [currentTheme, customColors])

  const toggleTheme = useCallback(() => {
    const newTheme = currentTheme === 'light' ? 'dark' : 'light'
    setCurrentTheme(newTheme)
    AsyncStorage.setItem('theme_preference', newTheme)
  }, [currentTheme])

  const setTheme = useCallback((themeName: string) => {
    setCurrentTheme(themeName)
    AsyncStorage.setItem('theme_preference', themeName)
  }, [])

  const contextValue = {
    theme,
    isDark: currentTheme === 'dark',
    toggleTheme,
    setTheme,
  }

  return (
    <ThemeContext.Provider value={contextValue}>
      <StatusBar
        barStyle={currentTheme === 'dark' ? 'light-content' : 'dark-content'}
        backgroundColor={theme.colors.surface}
      />
      {children}
    </ThemeContext.Provider>
  )
}

// 主题钩子
export const useTheme = () => {
  const context = useContext(ThemeContext)
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider')
  }
  return context
}
```

## 10. 结论

JobView Android UI/UX 设计规范基于 Material Design 3 指导原则，结合移动端用户体验最佳实践，提供了完整的设计系统和实施指南。通过统一的视觉语言、直观的交互模式、完善的可访问性支持和灵活的个性化选项，确保应用提供优质的用户体验。

### 关键设计特点：

1. **现代化视觉系统**: 基于 Material Design 3 的动态色彩和组件设计
2. **移动端优化**: 针对触摸操作、单手使用、多尺寸屏幕的专门优化
3. **一致性保证**: 统一的设计令牌和组件规范确保界面一致性
4. **可访问性优先**: 符合 WCAG 2.1 AA 标准的无障碍设计
5. **响应式适配**: 支持多种屏幕尺寸和方向的响应式布局
6. **个性化支持**: 丰富的主题和个性化定制选项

该设计规范为 JobView Android 应用提供了完整的用户界面设计基础，确保应用具备现代化、专业化、易用性的用户体验。