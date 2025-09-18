import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  KeyboardAvoidingView,
  Platform,
  Alert,
  ScrollView,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAppDispatch, useAppSelector, loginUser, setError } from '@store';

interface FormData {
  username: string;
  password: string;
}

const LoginScreen: React.FC = () => {
  const navigation = useNavigation();
  const dispatch = useAppDispatch();
  const { isLoading, error } = useAppSelector(state => state.auth);

  const [formData, setFormData] = useState<FormData>({
    username: '',
    password: '',
  });
  const [formErrors, setFormErrors] = useState<Partial<FormData>>({});

  // Clear error when component mounts
  useEffect(() => {
    dispatch(setError(null));
  }, [dispatch]);

  // Show error alert
  useEffect(() => {
    if (error) {
      Alert.alert('登录失败', error, [
        { text: '确定', onPress: () => dispatch(setError(null)) }
      ]);
    }
  }, [error, dispatch]);

  const validateForm = (): boolean => {
    const errors: Partial<FormData> = {};

    if (!formData.username.trim()) {
      errors.username = '请输入用户名或邮箱';
    }

    if (!formData.password.trim()) {
      errors.password = '请输入密码';
    } else if (formData.password.length < 6) {
      errors.password = '密码至少6位';
    }

    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleLogin = async () => {
    if (!validateForm()) {
      return;
    }

    try {
      const result = await dispatch(loginUser({
        username: formData.username.trim(),
        password: formData.password,
      }));

      if (loginUser.fulfilled.match(result)) {
        // Login successful, navigation will be handled by AppNavigator
        console.log('Login successful');
      }
    } catch (err) {
      console.error('Login error:', err);
    }
  };

  const handleRegister = () => {
    navigation.navigate('Register' as never);
  };

  const handleForgotPassword = () => {
    navigation.navigate('ForgotPassword' as never);
  };

  return (
    <KeyboardAvoidingView
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView contentContainerStyle={styles.scrollContent}>
        {/* Logo Section */}
        <View style={styles.logoSection}>
          <View style={styles.logo}>
            <Text style={styles.logoText}>JV</Text>
          </View>
          <Text style={styles.title}>欢迎回来</Text>
          <Text style={styles.subtitle}>登录您的 JobView 账户</Text>
        </View>

        {/* Login Form */}
        <View style={styles.form}>
          <View style={styles.inputGroup}>
            <Text style={styles.label}>用户名或邮箱</Text>
            <TextInput
              style={[styles.input, formErrors.username && styles.inputError]}
              placeholder="请输入用户名或邮箱"
              value={formData.username}
              onChangeText={(text) => {
                setFormData(prev => ({ ...prev, username: text }));
                if (formErrors.username) {
                  setFormErrors(prev => ({ ...prev, username: '' }));
                }
              }}
              autoCapitalize="none"
              autoCorrect={false}
              keyboardType="email-address"
              editable={!isLoading}
            />
            {formErrors.username && (
              <Text style={styles.errorText}>{formErrors.username}</Text>
            )}
          </View>

          <View style={styles.inputGroup}>
            <Text style={styles.label}>密码</Text>
            <TextInput
              style={[styles.input, formErrors.password && styles.inputError]}
              placeholder="请输入密码"
              secureTextEntry
              value={formData.password}
              onChangeText={(text) => {
                setFormData(prev => ({ ...prev, password: text }));
                if (formErrors.password) {
                  setFormErrors(prev => ({ ...prev, password: '' }));
                }
              }}
              editable={!isLoading}
            />
            {formErrors.password && (
              <Text style={styles.errorText}>{formErrors.password}</Text>
            )}
          </View>
        </View>

        {/* Action Buttons */}
        <View style={styles.actions}>
          <TouchableOpacity
            style={[
              styles.loginButton,
              (isLoading || !formData.username || !formData.password) && styles.loginButtonDisabled
            ]}
            onPress={handleLogin}
            disabled={isLoading || !formData.username || !formData.password}
          >
            <Text style={styles.loginButtonText}>
              {isLoading ? '登录中...' : '登录'}
            </Text>
          </TouchableOpacity>

          <View style={styles.linkContainer}>
            <TouchableOpacity
              style={styles.linkButton}
              onPress={handleRegister}
              disabled={isLoading}
            >
              <Text style={styles.linkText}>注册账户</Text>
            </TouchableOpacity>

            <TouchableOpacity
              style={styles.linkButton}
              onPress={handleForgotPassword}
              disabled={isLoading}
            >
              <Text style={styles.linkText}>忘记密码?</Text>
            </TouchableOpacity>
          </View>

          {/* Demo Login Button for Development */}
          {__DEV__ && (
            <TouchableOpacity
              style={styles.demoButton}
              onPress={() => {
                setFormData({
                  username: 'demo@jobview.com',
                  password: 'demo123'
                });
              }}
            >
              <Text style={styles.demoButtonText}>使用演示账号</Text>
            </TouchableOpacity>
          )}
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#FEF7FF',
  },
  scrollContent: {
    flexGrow: 1,
    justifyContent: 'center',
    padding: 16,
  },
  logoSection: {
    alignItems: 'center',
    marginBottom: 32,
  },
  logo: {
    width: 80,
    height: 80,
    borderRadius: 40,
    backgroundColor: '#6750A4',
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: 16,
  },
  logoText: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#FFFFFF',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#1C1B1F',
    marginBottom: 8,
  },
  subtitle: {
    fontSize: 16,
    color: '#49454F',
    textAlign: 'center',
  },
  form: {
    marginBottom: 24,
  },
  inputGroup: {
    marginBottom: 16,
  },
  label: {
    fontSize: 14,
    color: '#1C1B1F',
    fontWeight: '500',
    marginBottom: 8,
  },
  input: {
    backgroundColor: '#FFFFFF',
    borderColor: '#79747E',
    borderWidth: 1,
    borderRadius: 4,
    padding: 12,
    fontSize: 16,
    color: '#1C1B1F',
  },
  inputError: {
    borderColor: '#BA1A1A',
  },
  errorText: {
    fontSize: 12,
    color: '#BA1A1A',
    marginTop: 4,
  },
  actions: {
    gap: 16,
  },
  loginButton: {
    backgroundColor: '#6750A4',
    padding: 16,
    borderRadius: 4,
    alignItems: 'center',
  },
  loginButtonDisabled: {
    backgroundColor: '#79747E',
    opacity: 0.6,
  },
  loginButtonText: {
    color: '#FFFFFF',
    fontSize: 16,
    fontWeight: '600',
  },
  linkContainer: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  linkButton: {
    paddingVertical: 8,
    paddingHorizontal: 4,
  },
  linkText: {
    color: '#6750A4',
    fontSize: 14,
    fontWeight: '500',
  },
  demoButton: {
    backgroundColor: '#E8DEF8',
    padding: 12,
    borderRadius: 4,
    alignItems: 'center',
    marginTop: 16,
  },
  demoButtonText: {
    color: '#21005D',
    fontSize: 14,
    fontWeight: '500',
  },
});

export default LoginScreen;