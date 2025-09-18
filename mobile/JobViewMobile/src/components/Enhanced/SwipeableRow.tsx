import React, { useRef, useState } from 'react';
import {
  View,
  Text,
  Animated,
  PanResponder,
  TouchableOpacity,
  StyleSheet,
  Dimensions,
} from 'react-native';
import Icon from 'react-native-vector-icons/MaterialIcons';

const { width: SCREEN_WIDTH } = Dimensions.get('window');
const SWIPE_THRESHOLD = 80;
const ACTION_WIDTH = 80;

interface SwipeAction {
  id: string;
  title: string;
  icon: string;
  color: string;
  backgroundColor: string;
  onPress: () => void;
}

interface SwipeableRowProps {
  children: React.ReactNode;
  leftActions?: SwipeAction[];
  rightActions?: SwipeAction[];
  onSwipe?: (direction: 'left' | 'right', actionId: string) => void;
  style?: any;
}

const SwipeableRow: React.FC<SwipeableRowProps> = ({
  children,
  leftActions = [],
  rightActions = [],
  onSwipe,
  style,
}) => {
  const translateX = useRef(new Animated.Value(0)).current;
  const [gestureStarted, setGestureStarted] = useState(false);
  const [currentOffset, setCurrentOffset] = useState(0);

  const resetPosition = () => {
    Animated.spring(translateX, {
      toValue: 0,
      useNativeDriver: true,
      tension: 100,
      friction: 8,
    }).start();
    setCurrentOffset(0);
  };

  const animateToPosition = (position: number) => {
    Animated.spring(translateX, {
      toValue: position,
      useNativeDriver: true,
      tension: 100,
      friction: 8,
    }).start();
    setCurrentOffset(position);
  };

  const panResponder = useRef(
    PanResponder.create({
      onStartShouldSetPanResponder: () => true,
      onMoveShouldSetPanResponder: (evt, gestureState) => {
        return Math.abs(gestureState.dx) > Math.abs(gestureState.dy) && Math.abs(gestureState.dx) > 10;
      },
      onPanResponderGrant: () => {
        setGestureStarted(true);
      },
      onPanResponderMove: (evt, gestureState) => {
        const { dx } = gestureState;

        // Calculate the constrained translation
        const maxLeftSwipe = leftActions.length * ACTION_WIDTH;
        const maxRightSwipe = -rightActions.length * ACTION_WIDTH;

        let constrainedTranslation = dx;
        if (dx > maxLeftSwipe) {
          constrainedTranslation = maxLeftSwipe + (dx - maxLeftSwipe) * 0.1;
        } else if (dx < maxRightSwipe) {
          constrainedTranslation = maxRightSwipe + (dx - maxRightSwipe) * 0.1;
        }

        translateX.setValue(constrainedTranslation);
      },
      onPanResponderRelease: (evt, gestureState) => {
        const { dx, vx } = gestureState;
        setGestureStarted(false);

        // Determine if we should snap to an action or reset
        const threshold = SWIPE_THRESHOLD;

        if (dx > threshold && leftActions.length > 0) {
          // Swipe right - show left actions
          const actionIndex = Math.min(
            Math.floor(dx / ACTION_WIDTH),
            leftActions.length - 1
          );
          const targetPosition = (actionIndex + 1) * ACTION_WIDTH;
          animateToPosition(targetPosition);
        } else if (dx < -threshold && rightActions.length > 0) {
          // Swipe left - show right actions
          const actionIndex = Math.min(
            Math.floor(Math.abs(dx) / ACTION_WIDTH),
            rightActions.length - 1
          );
          const targetPosition = -(actionIndex + 1) * ACTION_WIDTH;
          animateToPosition(targetPosition);
        } else {
          // Reset to center
          resetPosition();
        }
      },
    })
  ).current;

  const handleActionPress = (action: SwipeAction, direction: 'left' | 'right') => {
    action.onPress();
    onSwipe?.(direction, action.id);
    resetPosition();
  };

  const renderActions = (actions: SwipeAction[], side: 'left' | 'right') => {
    if (actions.length === 0) return null;

    return (
      <View style={[styles.actionsContainer, styles[`${side}Actions`]]}>
        {actions.map((action, index) => (
          <TouchableOpacity
            key={action.id}
            style={[
              styles.actionButton,
              { backgroundColor: action.backgroundColor },
            ]}
            onPress={() => handleActionPress(action, side)}
            activeOpacity={0.7}
          >
            <Icon name={action.icon} size={24} color={action.color} />
            <Text style={[styles.actionText, { color: action.color }]}>
              {action.title}
            </Text>
          </TouchableOpacity>
        ))}
      </View>
    );
  };

  return (
    <View style={[styles.container, style]}>
      {renderActions(leftActions, 'left')}
      {renderActions(rightActions, 'right')}

      <Animated.View
        style={[
          styles.content,
          {
            transform: [{ translateX }],
          },
        ]}
        {...panResponder.panHandlers}
      >
        {children}
      </Animated.View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    position: 'relative',
    overflow: 'hidden',
  },
  content: {
    backgroundColor: '#FFFFFF',
    zIndex: 1,
  },
  actionsContainer: {
    position: 'absolute',
    top: 0,
    bottom: 0,
    flexDirection: 'row',
    alignItems: 'center',
  },
  leftActions: {
    left: 0,
  },
  rightActions: {
    right: 0,
  },
  actionButton: {
    width: ACTION_WIDTH,
    height: '100%',
    justifyContent: 'center',
    alignItems: 'center',
    paddingHorizontal: 8,
  },
  actionText: {
    fontSize: 10,
    fontWeight: '500',
    marginTop: 4,
    textAlign: 'center',
  },
});

export default SwipeableRow;