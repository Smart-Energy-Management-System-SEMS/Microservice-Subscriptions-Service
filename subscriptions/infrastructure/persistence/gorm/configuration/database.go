package configuration

import (
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	model2 "microservice-subscriptions-service/subscriptions/infrastructure/persistence/gorm/model"
)

func NewDatabase(databaseURL string) (*gorm.DB, error) {
	if databaseURL == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	return gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
}

func AutoMigrate(db *gorm.DB) error {
	models := []any{
		&model2.SubscriptionPlanModel{},
		&model2.PlanFeatureModel{},
		&model2.SubscriptionModel{},
	}

	for _, m := range models {
		if !db.Migrator().HasTable(m) {
			if err := db.Migrator().CreateTable(m); err != nil {
				return err
			}
		}
	}

	if err := ensureSubscriptionStatusConstraint(db); err != nil {
		return err
	}

	return nil
}

func ensureSubscriptionStatusConstraint(db *gorm.DB) error {
	normalizeSQL := `UPDATE subscriptions
		SET status = UPPER(TRIM(status))
		WHERE status IS NOT NULL`
	if err := db.Exec(normalizeSQL).Error; err != nil {
		return fmt.Errorf("normalize subscription status failed: %w", err)
	}

	dropOldChecksSQL := `
DO $$
DECLARE r RECORD;
BEGIN
	FOR r IN
		SELECT conname
		FROM pg_constraint
		WHERE conrelid = 'subscriptions'::regclass
		  AND contype = 'c'
		  AND pg_get_constraintdef(oid) ILIKE '%status%'
	LOOP
		EXECUTE format('ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS %I', r.conname);
	END LOOP;
END $$;`
	if err := db.Exec(dropOldChecksSQL).Error; err != nil {
		return fmt.Errorf("drop old status checks failed: %w", err)
	}

	addSQL := `ALTER TABLE subscriptions
		ADD CONSTRAINT chk_subscription_status
		CHECK (status IN ('ACTIVE','INACTIVE','CANCELLED','PENDING_RENEWAL','EXPIRED'))`
	if err := db.Exec(addSQL).Error; err != nil {
		return fmt.Errorf("create chk_subscription_status failed: %w", err)
	}

	return nil
}
