import React from 'react';
import {
  View,
  Text,
  StyleSheet,
  TouchableOpacity,
  Animated,
  Pressable,
} from 'react-native';

interface EnhancedCardProps {
  children: React.ReactNode;
  title?: string;
  subtitle?: string;
  onPress?: () => void;
  style?: any;
  animated?: boolean;
  elevation?: number;
  padding?: number;
  margin?: number;
  borderRadius?: number;
}

const EnhancedCard: React.FC<EnhancedCardProps> = ({
  children,
  title,
  subtitle,
  onPress,
  style,
  animated = true,
  elevation = 2,
  padding = 16,
  margin = 8,
  borderRadius = 12,
}) => {
  const scaleAnim = React.useRef(new Animated.Value(1)).current;
  const shadowAnim = React.useRef(new Animated.Value(elevation)).current;

  const handlePressIn = () => {
    if (animated && onPress) {
      Animated.parallel([
        Animated.spring(scaleAnim, {
          toValue: 0.98,
          useNativeDriver: true,
          tension: 150,
          friction: 4,
        }),
        Animated.timing(shadowAnim, {
          toValue: elevation + 2,
          duration: 150,
          useNativeDriver: false,
        }),
      ]).start();
    }
  };

  const handlePressOut = () => {
    if (animated && onPress) {
      Animated.parallel([
        Animated.spring(scaleAnim, {
          toValue: 1,
          useNativeDriver: true,
          tension: 150,
          friction: 4,
        }),
        Animated.timing(shadowAnim, {
          toValue: elevation,
          duration: 150,
          useNativeDriver: false,
        }),
      ]).start();
    }
  };

  const cardStyle = [
    styles.card,
    {
      padding,
      margin,
      borderRadius,
      elevation: animated ? undefined : elevation,
    },
    animated && {
      shadowOpacity: 0.1,
      shadowRadius: 4,
      shadowOffset: { width: 0, height: 2 },
    },
    style,
  ];

  const animatedStyle = animated
    ? {
        transform: [{ scale: scaleAnim }],
        elevation: shadowAnim,
      }
    : {};

  const CardContent = () => (
    <>
      {(title || subtitle) && (
        <View style={styles.header}>
          {title && <Text style={styles.title}>{title}</Text>}
          {subtitle && <Text style={styles.subtitle}>{subtitle}</Text>}
        </View>
      )}
      <View style={styles.content}>
        {children}
      </View>
    </>
  );

  if (onPress) {
    return (
      <Animated.View style={[cardStyle, animatedStyle]}>
        <Pressable
          onPress={onPress}
          onPressIn={handlePressIn}
          onPressOut={handlePressOut}
          style={styles.pressable}
          android_ripple={{
            color: '#6750A4',
            borderless: false,
            radius: 200,
          }}
        >
          <CardContent />
        </Pressable>
      </Animated.View>
    );
  }

  return (
    <Animated.View style={[cardStyle, animatedStyle]}>
      <CardContent />
    </Animated.View>
  );
};

const styles = StyleSheet.create({
  card: {
    backgroundColor: '#FFFFFF',
    shadowColor: '#000',
  },
  pressable: {
    borderRadius: 12,
  },
  header: {
    marginBottom: 12,
  },
  title: {
    fontSize: 18,
    fontWeight: '600',
    color: '#1C1B1F',
    marginBottom: 4,
  },
  subtitle: {
    fontSize: 14,
    color: '#49454F',
    lineHeight: 20,
  },
  content: {
    flex: 1,
  },
});

export default EnhancedCard;