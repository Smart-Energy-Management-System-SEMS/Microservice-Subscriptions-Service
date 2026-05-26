package configuration

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	appconfig "microservice-subscriptions-service/subscriptions/infrastructure/configuration"
	"microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/model"
)

type defaultPlanSeed struct {
	Name          string
	Description   string
	Price         float64
	Currency      string
	BillingPeriod string
	Features      []model.PlanFeatureModel
}

func SeedDefaultPlans(db *gorm.DB, cfg appconfig.AppConfig) error {
	plans := []defaultPlanSeed{
		{
			Name:          "Free",
			Description:   "Start monitoring at no cost",
			Price:         0,
			Currency:      "PEN",
			BillingPeriod: "monthly",
			Features: []model.PlanFeatureModel{
				feature("BASIC_DASHBOARD", "Basic energy dashboard", "enabled"),
				feature("CONSUMPTION_ALERTS", "Essential consumption alerts", "enabled"),
				feature("LINKED_DEVICES_LIMIT", "Linked devices limit", "3"),
				stripePriceFeature(cfg.StripePriceFree),
			},
		},
		{
			Name:          "Plus",
			Description:   "For active homes",
			Price:         15,
			Currency:      "PEN",
			BillingPeriod: "monthly",
			Features: []model.PlanFeatureModel{
				feature("FREE_INCLUDED", "Everything in Free", "enabled"),
				feature("DEVICE_ANALYTICS", "Detailed device analytics", "enabled"),
				feature("SAVING_RECOMMENDATIONS", "Personalized saving recommendations", "enabled"),
				feature("MONTHLY_REPORTS", "Monthly savings reports", "enabled"),
				stripePriceFeature(cfg.StripePricePlus),
			},
		},
		{
			Name:          "Pro",
			Description:   "Advanced control and insights",
			Price:         25,
			Currency:      "PEN",
			BillingPeriod: "monthly",
			Features: []model.PlanFeatureModel{
				feature("PLUS_INCLUDED", "Everything in Plus", "enabled"),
				feature("UNLIMITED_DEVICES", "Unlimited linked devices", "enabled"),
				feature("PRIORITY_SUPPORT", "Priority support", "enabled"),
				feature("ADVANCED_EXPORT", "Advanced report export", "enabled"),
				stripePriceFeature(cfg.StripePricePro),
			},
		},
	}

	for _, seed := range plans {
		if err := upsertPlan(db, seed); err != nil {
			return err
		}
	}

	return nil
}

func upsertPlan(db *gorm.DB, seed defaultPlanSeed) error {
	var plan model.SubscriptionPlanModel
	err := db.Where("name = ?", seed.Name).First(&plan).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}

	if err == gorm.ErrRecordNotFound {
		plan = model.SubscriptionPlanModel{
			PlanID:        uuid.NewString(),
			Name:          seed.Name,
			Description:   seed.Description,
			Price:         seed.Price,
			Currency:      seed.Currency,
			BillingPeriod: seed.BillingPeriod,
			Active:        true,
			CreatedAt:     time.Now().UTC(),
		}
		if createErr := db.Create(&plan).Error; createErr != nil {
			return createErr
		}
	}

	updates := map[string]any{
		"description":    seed.Description,
		"price":          seed.Price,
		"currency":       seed.Currency,
		"billing_period": seed.BillingPeriod,
		"active":         true,
	}
	if err = db.Model(&plan).Updates(updates).Error; err != nil {
		return err
	}

	for _, f := range seed.Features {
		if strings.TrimSpace(f.FeatureValue) == "" {
			continue
		}
		var existing model.PlanFeatureModel
		findErr := db.Where("plan_id = ? AND feature_code = ?", plan.PlanID, f.FeatureCode).First(&existing).Error
		if findErr != nil && findErr != gorm.ErrRecordNotFound {
			return findErr
		}
		if findErr == gorm.ErrRecordNotFound {
			f.FeatureID = uuid.NewString()
			f.PlanID = plan.PlanID
			f.CreatedAt = time.Now().UTC()
			if createErr := db.Create(&f).Error; createErr != nil {
				return createErr
			}
			continue
		}
		if err = db.Model(&existing).Updates(map[string]any{"feature_name": f.FeatureName, "feature_value": f.FeatureValue}).Error; err != nil {
			return err
		}
	}

	return nil
}

func feature(code, name, value string) model.PlanFeatureModel {
	return model.PlanFeatureModel{FeatureCode: code, FeatureName: name, FeatureValue: value}
}

func stripePriceFeature(priceID string) model.PlanFeatureModel {
	return feature("STRIPE_PRICE_ID", "Stripe Price ID", strings.TrimSpace(priceID))
}
