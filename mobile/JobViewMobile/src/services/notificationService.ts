import PushNotification from 'react-native-push-notification';
import PushNotificationIOS from '@react-native-community/push-notification-ios';
import { Platform, Alert, PermissionsAndroid } from 'react-native';
import AsyncStorage from '@react-native-async-storage/async-storage';
import { ApplicationReminder } from '@types';

export interface NotificationSettings {
  enabled: boolean;
  interviewReminders: boolean;
  followUpReminders: boolean;
  statusUpdates: boolean;
  sound: boolean;
  vibration: boolean;
  quietHours: {
    enabled: boolean;
    startHour: number;
    endHour: number;
  };
}

export interface ScheduledNotification {
  id: string;
  reminderId: string;
  title: string;
  message: string;
  date: Date;
  data?: any;
}

export class NotificationService {
  private static readonly STORAGE_KEY = '@jobview/notification_settings';
  private static readonly SCHEDULED_KEY = '@jobview/scheduled_notifications';
  private static isInitialized = false;

  /**
   * 初始化通知服务
   */
  static async initialize(): Promise<boolean> {
    if (this.isInitialized) return true;

    try {
      if (Platform.OS === 'android') {
        // 请求Android通知权限
        const granted = await PermissionsAndroid.request(
          PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS,
          {
            title: 'JobView通知权限',
            message: 'JobView需要通知权限来提醒您重要的求职事项',
            buttonNeutral: '稍后询问',
            buttonNegative: '拒绝',
            buttonPositive: '允许',
          }
        );

        if (granted !== PermissionsAndroid.RESULTS.GRANTED) {
          console.log('通知权限被拒绝');
          return false;
        }
      }

      // 配置PushNotification
      PushNotification.configure({
        onRegister: function (token: any) {
          console.log('FCM Token:', token);
        },

        onNotification: function (notification: any) {
          console.log('收到通知:', notification);

          if (notification.userInteraction) {
            // 用户点击了通知
            NotificationService.handleNotificationClick(notification);
          }

          // iOS需要调用completion handler
          if (Platform.OS === 'ios') {
            notification.finish(PushNotificationIOS.FetchResult.NoData);
          }
        },

        onAction: function (notification: any) {
          console.log('通知操作:', notification.action);
        },

        onRegistrationError: function (err: any) {
          console.error('通知注册失败:', err.message);
        },

        permissions: {
          alert: true,
          badge: true,
          sound: true,
        },

        popInitialNotification: true,
        requestPermissions: Platform.OS === 'ios',
      });

      // 创建Android通知频道
      if (Platform.OS === 'android') {
        PushNotification.createChannel(
          {
            channelId: 'jobview-reminders',
            channelName: 'JobView提醒',
            channelDescription: '求职投递提醒通知',
            importance: 4,
            vibrate: true,
          },
          (created: boolean) => console.log(`通知频道创建: ${created}`)
        );

        PushNotification.createChannel(
          {
            channelId: 'jobview-updates',
            channelName: 'JobView更新',
            channelDescription: '状态更新通知',
            importance: 3,
            vibrate: true,
          },
          (created: boolean) => console.log(`更新频道创建: ${created}`)
        );
      }

      this.isInitialized = true;
      return true;
    } catch (error) {
      console.error('通知服务初始化失败:', error);
      return false;
    }
  }

  /**
   * 获取通知设置
   */
  static async getSettings(): Promise<NotificationSettings> {
    try {
      const settingsStr = await AsyncStorage.getItem(this.STORAGE_KEY);
      if (settingsStr) {
        return JSON.parse(settingsStr);
      }
    } catch (error) {
      console.error('获取通知设置失败:', error);
    }

    // 返回默认设置
    return {
      enabled: true,
      interviewReminders: true,
      followUpReminders: true,
      statusUpdates: true,
      sound: true,
      vibration: true,
      quietHours: {
        enabled: false,
        startHour: 22,
        endHour: 8,
      },
    };
  }

  /**
   * 保存通知设置
   */
  static async saveSettings(settings: NotificationSettings): Promise<void> {
    try {
      await AsyncStorage.setItem(this.STORAGE_KEY, JSON.stringify(settings));
    } catch (error) {
      console.error('保存通知设置失败:', error);
      throw error;
    }
  }

  /**
   * 安排提醒通知
   */
  static async scheduleReminder(reminder: ApplicationReminder): Promise<boolean> {
    try {
      await this.initialize();

      const settings = await this.getSettings();
      if (!settings.enabled || !settings.interviewReminders) {
        return false;
      }

      const notificationDate = new Date(reminder.reminderDate);
      const now = new Date();

      // 检查是否已过期
      if (notificationDate <= now) {
        console.log('提醒时间已过期');
        return false;
      }

      // 检查静音时间
      if (this.isInQuietHours(notificationDate, settings.quietHours)) {
        // 调整到静音结束时间
        notificationDate.setHours(settings.quietHours.endHour, 0, 0, 0);
      }

      const notificationId = `reminder_${reminder.id}`;

      PushNotification.localNotificationSchedule({
        id: notificationId,
        channelId: 'jobview-reminders',
        title: reminder.title,
        message: reminder.description || '',
        date: notificationDate,
        soundName: settings.sound ? 'default' : undefined,
        vibrate: settings.vibration,
        playSound: settings.sound,
        userInfo: {
          type: 'reminder',
          reminderId: reminder.id,
          applicationId: reminder.applicationId,
        },
        actions: ['查看详情', '完成提醒'],
      });

      // 保存已安排的通知记录
      await this.saveScheduledNotification({
        id: notificationId,
        reminderId: reminder.id,
        title: reminder.title,
        message: reminder.description || '',
        date: notificationDate,
        data: {
          type: 'reminder',
          applicationId: reminder.applicationId,
        },
      });

      console.log(`提醒已安排: ${reminder.title} at ${notificationDate}`);
      return true;
    } catch (error) {
      console.error('安排提醒失败:', error);
      return false;
    }
  }

  /**
   * 取消提醒通知
   */
  static async cancelReminder(reminderId: string): Promise<void> {
    try {
      const notificationId = `reminder_${reminderId}`;

      PushNotification.cancelLocalNotification(notificationId);

      // 从已安排的通知记录中移除
      await this.removeScheduledNotification(notificationId);

      console.log(`提醒已取消: ${reminderId}`);
    } catch (error) {
      console.error('取消提醒失败:', error);
    }
  }

  /**
   * 发送即时通知
   */
  static async sendInstantNotification(
    title: string,
    message: string,
    data?: any
  ): Promise<void> {
    try {
      await this.initialize();

      const settings = await this.getSettings();
      if (!settings.enabled) {
        return;
      }

      PushNotification.localNotification({
        channelId: 'jobview-updates',
        title,
        message,
        soundName: settings.sound ? 'default' : undefined,
        vibrate: settings.vibration,
        playSound: settings.sound,
        userInfo: data,
      });
    } catch (error) {
      console.error('发送即时通知失败:', error);
    }
  }

  /**
   * 发送状态更新通知
   */
  static async sendStatusUpdateNotification(
    companyName: string,
    position: string,
    oldStatus: string,
    newStatus: string
  ): Promise<void> {
    const settings = await this.getSettings();
    if (!settings.statusUpdates) {
      return;
    }

    await this.sendInstantNotification(
      '状态更新',
      `${companyName} - ${position} 状态从"${oldStatus}"更新为"${newStatus}"`,
      {
        type: 'status_update',
        companyName,
        position,
        newStatus,
      }
    );
  }

  /**
   * 获取所有已安排的通知
   */
  static async getScheduledNotifications(): Promise<ScheduledNotification[]> {
    try {
      const notificationsStr = await AsyncStorage.getItem(this.SCHEDULED_KEY);
      if (notificationsStr) {
        const notifications = JSON.parse(notificationsStr);
        return notifications.map((n: any) => ({
          ...n,
          date: new Date(n.date),
        }));
      }
    } catch (error) {
      console.error('获取已安排通知失败:', error);
    }
    return [];
  }

  /**
   * 清理过期的通知
   */
  static async cleanupExpiredNotifications(): Promise<void> {
    try {
      const notifications = await this.getScheduledNotifications();
      const now = new Date();

      const validNotifications = notifications.filter(n => n.date > now);

      await AsyncStorage.setItem(this.SCHEDULED_KEY, JSON.stringify(validNotifications));

      console.log(`清理了 ${notifications.length - validNotifications.length} 个过期通知`);
    } catch (error) {
      console.error('清理过期通知失败:', error);
    }
  }

  /**
   * 处理通知点击
   */
  private static handleNotificationClick(notification: any): void {
    const { userInfo } = notification;

    if (userInfo?.type === 'reminder') {
      // 导航到相关的应用详情页面
      // 这里需要通过导航服务来处理
      console.log('点击提醒通知:', userInfo);
    } else if (userInfo?.type === 'status_update') {
      // 导航到应用列表或相关页面
      console.log('点击状态更新通知:', userInfo);
    }
  }

  /**
   * 检查是否在静音时间内
   */
  private static isInQuietHours(date: Date, quietHours: NotificationSettings['quietHours']): boolean {
    if (!quietHours.enabled) return false;

    const hour = date.getHours();
    const { startHour, endHour } = quietHours;

    if (startHour < endHour) {
      // 同一天内的静音时间，如 22:00 - 08:00 (跨天)
      return hour >= startHour || hour < endHour;
    } else {
      // 跨天的静音时间，如 22:00 - 08:00
      return hour >= startHour || hour < endHour;
    }
  }

  /**
   * 保存已安排的通知记录
   */
  private static async saveScheduledNotification(notification: ScheduledNotification): Promise<void> {
    try {
      const notifications = await this.getScheduledNotifications();
      notifications.push(notification);
      await AsyncStorage.setItem(this.SCHEDULED_KEY, JSON.stringify(notifications));
    } catch (error) {
      console.error('保存通知记录失败:', error);
    }
  }

  /**
   * 移除已安排的通知记录
   */
  private static async removeScheduledNotification(notificationId: string): Promise<void> {
    try {
      const notifications = await this.getScheduledNotifications();
      const filteredNotifications = notifications.filter(n => n.id !== notificationId);
      await AsyncStorage.setItem(this.SCHEDULED_KEY, JSON.stringify(filteredNotifications));
    } catch (error) {
      console.error('移除通知记录失败:', error);
    }
  }

  /**
   * 获取通知权限状态
   */
  static async getPermissionStatus(): Promise<'granted' | 'denied' | 'unknown'> {
    try {
      if (Platform.OS === 'android') {
        const granted = await PermissionsAndroid.check(
          PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS
        );
        return granted ? 'granted' : 'denied';
      } else {
        // iOS权限检查
        return new Promise<'granted' | 'denied'>((resolve) => {
          PushNotificationIOS.checkPermissions((permissions) => {
            if (permissions.alert || permissions.badge || permissions.sound) {
              resolve('granted');
            } else {
              resolve('denied');
            }
          });
        });
      }
    } catch (error) {
      console.error('检查通知权限失败:', error);
      return 'unknown';
    }
  }

  /**
   * 请求通知权限
   */
  static async requestPermissions(): Promise<boolean> {
    try {
      if (Platform.OS === 'android') {
        const granted = await PermissionsAndroid.request(
          PermissionsAndroid.PERMISSIONS.POST_NOTIFICATIONS,
          {
            title: 'JobView通知权限',
            message: 'JobView需要通知权限来提醒您重要的求职事项',
            buttonPositive: '允许',
          }
        );
        return granted === PermissionsAndroid.RESULTS.GRANTED;
      } else {
        return new Promise<boolean>((resolve) => {
          PushNotificationIOS.requestPermissions({
            alert: true,
            badge: true,
            sound: true,
          }).then((permissions) => {
            resolve(!!(permissions.alert || permissions.badge || permissions.sound));
          });
        });
      }
    } catch (error) {
      console.error('请求通知权限失败:', error);
      return false;
    }
  }

  /**
   * 显示权限设置提示
   */
  static showPermissionAlert(): void {
    Alert.alert(
      '通知权限',
      '为了及时提醒您重要的求职事项，请在设置中允许JobView发送通知。',
      [
        { text: '稍后设置', style: 'cancel' },
        { text: '去设置', onPress: () => {
          // 这里可以打开系统设置页面
          // 需要安装 react-native-settings 或类似库
        }}
      ]
    );
  }

  /**
   * 清除所有通知
   */
  static async clearAllNotifications(): Promise<void> {
    try {
      PushNotification.cancelAllLocalNotifications();
      await AsyncStorage.removeItem(this.SCHEDULED_KEY);
      console.log('所有通知已清除');
    } catch (error) {
      console.error('清除通知失败:', error);
    }
  }

  /**
   * 获取通知统计
   */
  static async getNotificationStats(): Promise<{
    totalScheduled: number;
    upcomingToday: number;
    upcomingWeek: number;
    expired: number;
  }> {
    try {
      const notifications = await this.getScheduledNotifications();
      const now = new Date();
      const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
      const tomorrow = new Date(today.getTime() + 24 * 60 * 60 * 1000);
      const nextWeek = new Date(today.getTime() + 7 * 24 * 60 * 60 * 1000);

      return {
        totalScheduled: notifications.length,
        upcomingToday: notifications.filter(n => n.date >= today && n.date < tomorrow).length,
        upcomingWeek: notifications.filter(n => n.date >= today && n.date < nextWeek).length,
        expired: notifications.filter(n => n.date < now).length,
      };
    } catch (error) {
      console.error('获取通知统计失败:', error);
      return {
        totalScheduled: 0,
        upcomingToday: 0,
        upcomingWeek: 0,
        expired: 0,
      };
    }
  }
}