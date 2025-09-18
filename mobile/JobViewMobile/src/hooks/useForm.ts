import { useState, useCallback, useMemo } from 'react';

type ValidationRule<T> = (value: T) => string | null;

interface FieldConfig<T = any> {
  initialValue: T;
  rules?: ValidationRule<T>[];
  required?: boolean;
  requiredMessage?: string;
}

interface FormConfig {
  [fieldName: string]: FieldConfig;
}

interface FieldState<T = any> {
  value: T;
  error: string | null;
  touched: boolean;
  dirty: boolean;
}

interface FormState {
  [fieldName: string]: FieldState;
}

// Common validation rules
export const validationRules = {
  required: (message = '此字段为必填项'): ValidationRule<any> => (value) => {
    if (value === null || value === undefined || value === '' ||
        (Array.isArray(value) && value.length === 0)) {
      return message;
    }
    return null;
  },

  minLength: (min: number, message?: string): ValidationRule<string> => (value) => {
    if (value && value.length < min) {
      return message || `最少需要 ${min} 个字符`;
    }
    return null;
  },

  maxLength: (max: number, message?: string): ValidationRule<string> => (value) => {
    if (value && value.length > max) {
      return message || `最多允许 ${max} 个字符`;
    }
    return null;
  },

  email: (message = '请输入有效的邮箱地址'): ValidationRule<string> => (value) => {
    if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      return message;
    }
    return null;
  },

  phone: (message = '请输入有效的手机号码'): ValidationRule<string> => (value) => {
    if (value && !/^1[3-9]\d{9}$/.test(value)) {
      return message;
    }
    return null;
  },

  url: (message = '请输入有效的URL'): ValidationRule<string> => (value) => {
    if (value && !/^https?:\/\/.+/.test(value)) {
      return message;
    }
    return null;
  },

  number: (message = '请输入有效的数字'): ValidationRule<string> => (value) => {
    if (value && isNaN(Number(value))) {
      return message;
    }
    return null;
  },

  pattern: (pattern: RegExp, message: string): ValidationRule<string> => (value) => {
    if (value && !pattern.test(value)) {
      return message;
    }
    return null;
  },

  custom: <T>(validator: (value: T) => boolean, message: string): ValidationRule<T> => (value) => {
    if (!validator(value)) {
      return message;
    }
    return null;
  },
};

export const useForm = <T extends FormConfig>(config: T) => {
  // Initialize form state
  const initialState = useMemo(() => {
    const state: FormState = {};
    Object.keys(config).forEach(fieldName => {
      const fieldConfig = config[fieldName];
      state[fieldName] = {
        value: fieldConfig.initialValue,
        error: null,
        touched: false,
        dirty: false,
      };
    });
    return state;
  }, [config]);

  const [formState, setFormState] = useState<FormState>(initialState);

  // Validate a single field
  const validateField = useCallback((fieldName: string, value: any): string | null => {
    const fieldConfig = config[fieldName];
    if (!fieldConfig) return null;

    // Check required
    if (fieldConfig.required) {
      const requiredError = validationRules.required(fieldConfig.requiredMessage)(value);
      if (requiredError) return requiredError;
    }

    // Check other rules
    if (fieldConfig.rules) {
      for (const rule of fieldConfig.rules) {
        const error = rule(value);
        if (error) return error;
      }
    }

    return null;
  }, [config]);

  // Set field value and validate
  const setFieldValue = useCallback((fieldName: string, value: any) => {
    setFormState(prev => {
      const currentField = prev[fieldName];
      const error = validateField(fieldName, value);

      return {
        ...prev,
        [fieldName]: {
          ...currentField,
          value,
          error,
          dirty: value !== config[fieldName].initialValue,
        },
      };
    });
  }, [validateField, config]);

  // Set field as touched
  const setFieldTouched = useCallback((fieldName: string, touched = true) => {
    setFormState(prev => ({
      ...prev,
      [fieldName]: {
        ...prev[fieldName],
        touched,
      },
    }));
  }, []);

  // Set field error manually
  const setFieldError = useCallback((fieldName: string, error: string | null) => {
    setFormState(prev => ({
      ...prev,
      [fieldName]: {
        ...prev[fieldName],
        error,
      },
    }));
  }, []);

  // Validate all fields
  const validateForm = useCallback(() => {
    const errors: { [key: string]: string } = {};
    let isValid = true;

    Object.keys(config).forEach(fieldName => {
      const error = validateField(fieldName, formState[fieldName].value);
      if (error) {
        errors[fieldName] = error;
        isValid = false;
      }
    });

    // Update all field errors
    setFormState(prev => {
      const newState = { ...prev };
      Object.keys(config).forEach(fieldName => {
        newState[fieldName] = {
          ...newState[fieldName],
          error: errors[fieldName] || null,
          touched: true,
        };
      });
      return newState;
    });

    return { isValid, errors };
  }, [config, formState, validateField]);

  // Reset form
  const resetForm = useCallback(() => {
    setFormState(initialState);
  }, [initialState]);

  // Get form values
  const getValues = useCallback(() => {
    const values: { [key: string]: any } = {};
    Object.keys(formState).forEach(fieldName => {
      values[fieldName] = formState[fieldName].value;
    });
    return values;
  }, [formState]);

  // Get field helpers
  const getFieldProps = useCallback((fieldName: string) => {
    const field = formState[fieldName];
    return {
      value: field.value,
      error: field.touched ? field.error : null,
      onChangeText: (value: any) => setFieldValue(fieldName, value),
      onBlur: () => setFieldTouched(fieldName, true),
    };
  }, [formState, setFieldValue, setFieldTouched]);

  // Check if form is valid
  const isValid = useMemo(() => {
    return Object.values(formState).every(field => !field.error);
  }, [formState]);

  // Check if form is dirty
  const isDirty = useMemo(() => {
    return Object.values(formState).some(field => field.dirty);
  }, [formState]);

  // Check if form has been touched
  const isTouched = useMemo(() => {
    return Object.values(formState).some(field => field.touched);
  }, [formState]);

  return {
    // State
    formState,
    isValid,
    isDirty,
    isTouched,

    // Actions
    setFieldValue,
    setFieldTouched,
    setFieldError,
    validateForm,
    resetForm,
    getValues,
    getFieldProps,

    // Field helpers
    fields: Object.keys(config).reduce((acc, fieldName) => {
      acc[fieldName] = getFieldProps(fieldName);
      return acc;
    }, {} as any),
  };
};