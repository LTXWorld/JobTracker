import React from 'react';
import { View, Text, StyleSheet } from 'react-native';
import { ApplicationStatus } from '../../types';

interface StatusBadgeProps {
  status: ApplicationStatus;
  size?: 'small' | 'medium' | 'large';
  variant?: 'filled' | 'outlined' | 'soft';
}

const StatusBadge: React.FC<StatusBadgeProps> = ({
  status,
  size = 'medium',
  variant = 'soft'
}) => {
  const getStatusConfig = (status: ApplicationStatus) => {
    const configs = {
      '已保存': { color: '#79747E', bgColor: '#F3F0F4', label: '已保存' },
      '已投递': { color: '#1976D2', bgColor: '#E3F2FD', label: '已投递' },
      '简历筛选中': { color: '#7B1FA2', bgColor: '#F3E5F5', label: '简历筛选中' },
      '笔试中': { color: '#F57C00', bgColor: '#FFF3E0', label: '笔试中' },
      '一面中': { color: '#388E3C', bgColor: '#E8F5E8', label: '一面中' },
      '二面中': { color: '#388E3C', bgColor: '#E8F5E8', label: '二面中' },
      '三面中': { color: '#388E3C', bgColor: '#E8F5E8', label: '三面中' },
      'HR面中': { color: '#388E3C', bgColor: '#E8F5E8', label: 'HR面中' },
      '等待offer': { color: '#6750A4', bgColor: '#F3E5F5', label: '等待offer' },
      '已收到offer': { color: '#2E7D32', bgColor: '#E8F5E8', label: '已收到offer' },
      '已拒绝offer': { color: '#B71C1C', bgColor: '#FFEBEE', label: '已拒绝offer' },
      '简历挂': { color: '#D32F2F', bgColor: '#FFEBEE', label: '简历挂' },
      '笔试挂': { color: '#D32F2F', bgColor: '#FFEBEE', label: '笔试挂' },
      '一面挂': { color: '#D32F2F', bgColor: '#FFEBEE', label: '一面挂' },
      '二面挂': { color: '#D32F2F', bgColor: '#FFEBEE', label: '二面挂' },
      '三面挂': { color: '#D32F2F', bgColor: '#FFEBEE', label: '三面挂' },
      'HR面挂': { color: '#D32F2F', bgColor: '#FFEBEE', label: 'HR面挂' },
      '被拒绝': { color: '#D32F2F', bgColor: '#FFEBEE', label: '被拒绝' },
      '已入职': { color: '#1B5E20', bgColor: '#E8F5E8', label: '已入职' },
      '已放弃': { color: '#424242', bgColor: '#F5F5F5', label: '已放弃' },
    };

    return configs[status] || { color: '#79747E', bgColor: '#F3F0F4', label: status };
  };

  const config = getStatusConfig(status);

  const getBadgeStyle = () => {
    const baseStyle = [styles.badge, styles[size]];

    switch (variant) {
      case 'filled':
        return [
          ...baseStyle,
          {
            backgroundColor: config.color,
            borderColor: config.color,
          },
        ];
      case 'outlined':
        return [
          ...baseStyle,
          styles.outlined,
          {
            borderColor: config.color,
            backgroundColor: 'transparent',
          },
        ];
      case 'soft':
      default:
        return [
          ...baseStyle,
          {
            backgroundColor: config.bgColor,
            borderColor: config.bgColor,
          },
        ];
    }
  };

  const getTextStyle = () => {
    const baseStyle = [styles.text, styles[`${size}Text`]];

    switch (variant) {
      case 'filled':
        return [...baseStyle, { color: '#FFFFFF' }];
      case 'outlined':
      case 'soft':
      default:
        return [...baseStyle, { color: config.color }];
    }
  };

  return (
    <View style={getBadgeStyle()}>
      <Text style={getTextStyle()}>{config.label}</Text>
    </View>
  );
};

const styles = StyleSheet.create({
  badge: {
    borderRadius: 100,
    borderWidth: 1,
    alignItems: 'center',
    justifyContent: 'center',
    alignSelf: 'flex-start',
  },
  outlined: {
    borderWidth: 1,
  },

  // Size variants
  small: {
    paddingHorizontal: 8,
    paddingVertical: 2,
    minHeight: 20,
  },
  medium: {
    paddingHorizontal: 12,
    paddingVertical: 4,
    minHeight: 24,
  },
  large: {
    paddingHorizontal: 16,
    paddingVertical: 6,
    minHeight: 32,
  },

  // Text sizes
  text: {
    fontWeight: '500',
    textAlign: 'center',
  },
  smallText: {
    fontSize: 10,
  },
  mediumText: {
    fontSize: 12,
  },
  largeText: {
    fontSize: 14,
  },
});

export default StatusBadge;