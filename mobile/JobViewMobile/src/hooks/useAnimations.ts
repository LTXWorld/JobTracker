import { useRef, useEffect } from 'react';
import { Animated, Easing } from 'react-native';

// Fade in animation hook
export const useFadeIn = (duration: number = 300, delay: number = 0) => {
  const opacity = useRef(new Animated.Value(0)).current;

  useEffect(() => {
    const timer = setTimeout(() => {
      Animated.timing(opacity, {
        toValue: 1,
        duration,
        easing: Easing.out(Easing.quad),
        useNativeDriver: true,
      }).start();
    }, delay);

    return () => clearTimeout(timer);
  }, [opacity, duration, delay]);

  return opacity;
};

// Scale animation hook
export const useScale = (duration: number = 300, initialScale: number = 0.8) => {
  const scale = useRef(new Animated.Value(initialScale)).current;

  useEffect(() => {
    Animated.spring(scale, {
      toValue: 1,
      duration,
      useNativeDriver: true,
      tension: 80,
      friction: 7,
    }).start();
  }, [scale, duration]);

  return scale;
};

// Slide in from bottom animation
export const useSlideInFromBottom = (duration: number = 400, distance: number = 100) => {
  const translateY = useRef(new Animated.Value(distance)).current;

  useEffect(() => {
    Animated.timing(translateY, {
      toValue: 0,
      duration,
      easing: Easing.out(Easing.back(1.2)),
      useNativeDriver: true,
    }).start();
  }, [translateY, duration, distance]);

  return translateY;
};

// Slide in from right animation
export const useSlideInFromRight = (duration: number = 350, distance: number = 100) => {
  const translateX = useRef(new Animated.Value(distance)).current;

  useEffect(() => {
    Animated.timing(translateX, {
      toValue: 0,
      duration,
      easing: Easing.out(Easing.quad),
      useNativeDriver: true,
    }).start();
  }, [translateX, duration, distance]);

  return translateX;
};

// Bounce animation hook
export const useBounce = (duration: number = 1000, scale: number = 0.1) => {
  const bounceValue = useRef(new Animated.Value(1)).current;

  const startBounce = () => {
    Animated.sequence([
      Animated.timing(bounceValue, {
        toValue: 1 + scale,
        duration: duration / 4,
        easing: Easing.out(Easing.quad),
        useNativeDriver: true,
      }),
      Animated.timing(bounceValue, {
        toValue: 1,
        duration: duration / 4,
        easing: Easing.in(Easing.quad),
        useNativeDriver: true,
      }),
    ]).start();
  };

  return { bounceValue, startBounce };
};

// Shake animation hook
export const useShake = (duration: number = 800) => {
  const shakeValue = useRef(new Animated.Value(0)).current;

  const startShake = () => {
    const shakeAnimation = Animated.sequence([
      Animated.timing(shakeValue, { toValue: 10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: -10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: 10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: -10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: 10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: -10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: 10, duration: duration / 8, useNativeDriver: true }),
      Animated.timing(shakeValue, { toValue: 0, duration: duration / 8, useNativeDriver: true }),
    ]);

    shakeAnimation.start();
  };

  return { shakeValue, startShake };
};

// Pulse animation hook
export const usePulse = (minScale: number = 0.95, maxScale: number = 1.05, duration: number = 1000) => {
  const pulseValue = useRef(new Animated.Value(minScale)).current;

  useEffect(() => {
    const pulse = Animated.loop(
      Animated.sequence([
        Animated.timing(pulseValue, {
          toValue: maxScale,
          duration: duration / 2,
          easing: Easing.inOut(Easing.quad),
          useNativeDriver: true,
        }),
        Animated.timing(pulseValue, {
          toValue: minScale,
          duration: duration / 2,
          easing: Easing.inOut(Easing.quad),
          useNativeDriver: true,
        }),
      ])
    );

    pulse.start();

    return () => pulse.stop();
  }, [pulseValue, minScale, maxScale, duration]);

  return pulseValue;
};

// Staggered animation hook for lists
export const useStaggeredAnimation = (itemCount: number, staggerDelay: number = 100) => {
  const animations = useRef(
    Array.from({ length: itemCount }, () => new Animated.Value(0))
  ).current;

  const startStaggered = () => {
    const staggeredAnimations = animations.map((anim, index) =>
      Animated.timing(anim, {
        toValue: 1,
        duration: 300,
        delay: index * staggerDelay,
        easing: Easing.out(Easing.quad),
        useNativeDriver: true,
      })
    );

    Animated.parallel(staggeredAnimations).start();
  };

  const resetAnimations = () => {
    animations.forEach(anim => anim.setValue(0));
  };

  return { animations, startStaggered, resetAnimations };
};

// Combined entrance animation
export const useEntranceAnimation = (delay: number = 0) => {
  const opacity = useRef(new Animated.Value(0)).current;
  const translateY = useRef(new Animated.Value(30)).current;
  const scale = useRef(new Animated.Value(0.9)).current;

  useEffect(() => {
    const timer = setTimeout(() => {
      Animated.parallel([
        Animated.timing(opacity, {
          toValue: 1,
          duration: 400,
          easing: Easing.out(Easing.quad),
          useNativeDriver: true,
        }),
        Animated.timing(translateY, {
          toValue: 0,
          duration: 400,
          easing: Easing.out(Easing.back(1.1)),
          useNativeDriver: true,
        }),
        Animated.spring(scale, {
          toValue: 1,
          tension: 80,
          friction: 7,
          useNativeDriver: true,
        }),
      ]).start();
    }, delay);

    return () => clearTimeout(timer);
  }, [opacity, translateY, scale, delay]);

  return {
    opacity,
    transform: [{ translateY }, { scale }],
  };
};